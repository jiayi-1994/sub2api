//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type openAIImagesStaticUpstream struct {
	HTTPUpstream
	status int
	body   string
}

func (u openAIImagesStaticUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	return &http.Response{
		StatusCode: u.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(u.body)),
	}, nil
}

func forwardImagesWithStaticUpstream(t *testing.T, status int, body string) (*OpenAIForwardResult, error, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{httpUpstream: openAIImagesStaticUpstream{status: status, body: body}, cfg: &config.Config{}}
	account := &Account{
		ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true,
		Credentials: map[string]any{"api_key": "test-key", "base_url": "https://images.example.test"},
	}
	reqBody := []byte(`{"model":"gpt-image-2.6-01","prompt":"cat","n":1}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(reqBody))
	parsed := &OpenAIImagesRequest{Model: "gpt-image-2.6-01", Endpoint: openAIImagesGenerationsEndpoint, ContentType: "application/json", N: 1}
	result, err := svc.ForwardImages(context.Background(), c, account, reqBody, parsed, "")
	return result, err, recorder
}

// 上游出图超时会回 202 + 未完成任务体（无图片）。必须在写回客户端前转成错误：
// 否则非流式计费会回退到请求 n=1，客户被扣一张却拿不到图。
func TestForwardImages_Accepted202WithoutImagesIsFailoverNotBilled(t *testing.T) {
	result, err, recorder := forwardImagesWithStaticUpstream(t, http.StatusAccepted,
		`{"request_id":"r1","status":"in_progress","stage":"generating","created":1,"data":[]}`)
	require.Nil(t, result)
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.True(t, failoverErr.RequestScopedTransient, "must not penalize the account")
	require.False(t, recorder.Body.Len() > 0, "nothing may be written before failover")
}

func TestForwardImages_Accepted202WithImageStillSucceeds(t *testing.T) {
	result, err, recorder := forwardImagesWithStaticUpstream(t, http.StatusAccepted,
		`{"status":"completed","stage":"completed","created":1,"data":[{"url":"https://img.example.test/a.png"}],"usage":{"input_tokens":11,"output_tokens":425}}`)
	require.NoError(t, err)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, http.StatusAccepted, recorder.Code)
}
