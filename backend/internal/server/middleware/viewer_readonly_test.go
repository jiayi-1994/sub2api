//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// viewerCase 定义 ViewerReadOnly 中间件的一条测试用例。
type viewerCase struct {
	name       string
	role       string
	method     string
	pattern    string // gin 路由模式（含 /api/v1 前缀）
	wantStatus int
}

func TestViewerReadOnlyAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []viewerCase{
		// viewer 允许的读路径
		{"viewer-dashboard-stats", service.RoleViewer, "GET", "/api/v1/admin/dashboard/stats", http.StatusOK},
		{"viewer-dashboard-snapshot", service.RoleViewer, "GET", "/api/v1/admin/dashboard/snapshot-v2", http.StatusOK},
		{"viewer-dashboard-users-usage-post", service.RoleViewer, "POST", "/api/v1/admin/dashboard/users-usage", http.StatusOK},
		{"viewer-usage-list", service.RoleViewer, "GET", "/api/v1/admin/usage", http.StatusOK},
		{"viewer-usage-stats", service.RoleViewer, "GET", "/api/v1/admin/usage/stats", http.StatusOK},
		{"viewer-usage-cleanup-list", service.RoleViewer, "GET", "/api/v1/admin/usage/cleanup-tasks", http.StatusOK},

		// viewer 禁止的写路径
		{"viewer-aggregation-backfill", service.RoleViewer, "POST", "/api/v1/admin/dashboard/aggregation/backfill", http.StatusForbidden},
		{"viewer-cleanup-create", service.RoleViewer, "POST", "/api/v1/admin/usage/cleanup-tasks", http.StatusForbidden},
		{"viewer-cleanup-cancel", service.RoleViewer, "POST", "/api/v1/admin/usage/cleanup-tasks/:id/cancel", http.StatusForbidden},
		{"viewer-users-list", service.RoleViewer, "GET", "/api/v1/admin/users", http.StatusForbidden},
		{"viewer-settings", service.RoleViewer, "GET", "/api/v1/admin/settings", http.StatusForbidden},
		{"viewer-payment-dashboard", service.RoleViewer, "GET", "/api/v1/admin/payment/dashboard", http.StatusForbidden},

		// admin 应当全部放行
		{"admin-dashboard-stats", service.RoleAdmin, "GET", "/api/v1/admin/dashboard/stats", http.StatusOK},
		{"admin-aggregation-backfill", service.RoleAdmin, "POST", "/api/v1/admin/dashboard/aggregation/backfill", http.StatusOK},
		{"admin-users-list", service.RoleAdmin, "GET", "/api/v1/admin/users", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set(string(ContextKeyUserRole), tc.role)
				c.Next()
			})
			r.Use(ViewerReadOnly())
			r.Handle(tc.method, tc.pattern, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tc.method, tc.pattern, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code, "role=%s %s %s", tc.role, tc.method, tc.pattern)
		})
	}
}
