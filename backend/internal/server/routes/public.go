package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterPublicRoutes 注册无需认证的公开路由（如 /monitoring 状态页 API）。
// 所有公开端点应：fail-open 限流（Redis 故障时不阻断只读状态页）+ Cache-Control。
func RegisterPublicRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	redisClient *redis.Client,
) {
	rateLimiter := middleware.NewRateLimiter(redisClient)

	public := v1.Group("/public")
	{
		// 公开渠道监控（未登录访客可查看）
		monitoring := public.Group("/monitoring")
		monitoring.Use(rateLimiter.LimitWithOptions(
			"public-monitoring",
			60,
			time.Minute,
			middleware.RateLimitOptions{FailureMode: middleware.RateLimitFailOpen},
		))
		{
			monitoring.GET("", h.ChannelMonitor.ListPublic)
			monitoring.GET("/:id/status", h.ChannelMonitor.GetStatusPublic)
		}
	}
}
