package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// 公开渠道监控 handler 提供未登录访客只读视图。
// 复用 ChannelMonitorUserHandler 的 service 依赖；通过 PublicEnabled 开关单独控制。
//
// 数据脱敏：
//   - 不返回 group_name（防止内部路由策略泄露）
//   - 保留 provider、name、各模型可用率/延迟（与 coderelay.cn/monitoring 一致）

// channelMonitorPublicCacheTTL 控制公开 List 接口的进程内缓存窗口。
// 与前端 Cache-Control: 30s 对齐，避免高并发刷新打到 DB。
const channelMonitorPublicCacheTTL = 30 * time.Second

type channelMonitorPublicListCache struct {
	mu      sync.RWMutex
	items   []channelMonitorPublicListItem
	expires time.Time
}

func (c *channelMonitorPublicListCache) get() ([]channelMonitorPublicListItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().Before(c.expires) {
		return c.items, true
	}
	return nil, false
}

func (c *channelMonitorPublicListCache) set(items []channelMonitorPublicListItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = items
	c.expires = time.Now().Add(channelMonitorPublicCacheTTL)
}

var publicListCache = &channelMonitorPublicListCache{}

// channelMonitorPublicListItem 公开列表项：去掉 group_name + api_key_decrypt_failed 等内部字段。
type channelMonitorPublicListItem struct {
	ID                   int64                                `json:"id"`
	Name                 string                               `json:"name"`
	Provider             string                               `json:"provider"`
	PrimaryModel         string                               `json:"primary_model"`
	PrimaryStatus        string                               `json:"primary_status"`
	PrimaryLatencyMs     *int                                 `json:"primary_latency_ms"`
	PrimaryPingLatencyMs *int                                 `json:"primary_ping_latency_ms"`
	Availability7d       float64                              `json:"availability_7d"`
	ExtraModels          []dto.ChannelMonitorExtraModelStatus `json:"extra_models"`
	Timeline             []channelMonitorUserTimelinePoint    `json:"timeline"`
}

// channelMonitorPublicDetailResponse 公开详情：去掉 group_name。
type channelMonitorPublicDetailResponse struct {
	ID       int64                         `json:"id"`
	Name     string                        `json:"name"`
	Provider string                        `json:"provider"`
	Models   []channelMonitorUserModelStat `json:"models"`
}

func userMonitorViewToPublicItem(v *service.UserMonitorView) channelMonitorPublicListItem {
	extras := make([]dto.ChannelMonitorExtraModelStatus, 0, len(v.ExtraModels))
	for _, e := range v.ExtraModels {
		extras = append(extras, dto.ChannelMonitorExtraModelStatus{
			Model:     e.Model,
			Status:    e.Status,
			LatencyMs: e.LatencyMs,
		})
	}
	timeline := make([]channelMonitorUserTimelinePoint, 0, len(v.Timeline))
	for _, p := range v.Timeline {
		timeline = append(timeline, channelMonitorUserTimelinePoint{
			Status:        p.Status,
			LatencyMs:     p.LatencyMs,
			PingLatencyMs: p.PingLatencyMs,
			CheckedAt:     p.CheckedAt.UTC().Format(time.RFC3339),
		})
	}
	return channelMonitorPublicListItem{
		ID:                   v.ID,
		Name:                 v.Name,
		Provider:             v.Provider,
		PrimaryModel:         v.PrimaryModel,
		PrimaryStatus:        v.PrimaryStatus,
		PrimaryLatencyMs:     v.PrimaryLatencyMs,
		PrimaryPingLatencyMs: v.PrimaryPingLatencyMs,
		Availability7d:       v.Availability7d,
		ExtraModels:          extras,
		Timeline:             timeline,
	}
}

func userMonitorDetailToPublicResponse(d *service.UserMonitorDetail) *channelMonitorPublicDetailResponse {
	models := make([]channelMonitorUserModelStat, 0, len(d.Models))
	for _, m := range d.Models {
		models = append(models, channelMonitorUserModelStat{
			Model:           m.Model,
			LatestStatus:    m.LatestStatus,
			LatestLatencyMs: m.LatestLatencyMs,
			Availability7d:  m.Availability7d,
			Availability15d: m.Availability15d,
			Availability30d: m.Availability30d,
			AvgLatency7dMs:  m.AvgLatency7dMs,
		})
	}
	return &channelMonitorPublicDetailResponse{
		ID:       d.ID,
		Name:     d.Name,
		Provider: d.Provider,
		Models:   models,
	}
}

// publicFeatureEnabled 公开页要求 Enabled && PublicEnabled 同时为真。
func (h *ChannelMonitorUserHandler) publicFeatureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return true
	}
	rt := h.settingService.GetChannelMonitorRuntime(c.Request.Context())
	return rt.Enabled && rt.PublicEnabled
}

// ListPublic GET /api/v1/public/monitoring
func (h *ChannelMonitorUserHandler) ListPublic(c *gin.Context) {
	if !h.publicFeatureEnabled(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=30")

	if cached, ok := publicListCache.get(); ok {
		response.Success(c, gin.H{"items": cached})
		return
	}

	views, err := h.monitorService.ListUserView(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items := make([]channelMonitorPublicListItem, 0, len(views))
	for _, v := range views {
		items = append(items, userMonitorViewToPublicItem(v))
	}
	publicListCache.set(items)
	response.Success(c, gin.H{"items": items})
}

// GetStatusPublic GET /api/v1/public/monitoring/:id/status
func (h *ChannelMonitorUserHandler) GetStatusPublic(c *gin.Context) {
	if !h.publicFeatureEnabled(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=30")

	id, ok := admin.ParseChannelMonitorID(c)
	if !ok {
		return
	}
	detail, err := h.monitorService.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, userMonitorDetailToPublicResponse(detail))
}
