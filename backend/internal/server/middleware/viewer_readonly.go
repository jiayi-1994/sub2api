package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// viewerAllowed 是 role=viewer（只读管理员）可访问的 (method, gin pattern) 白名单。
// 使用 c.FullPath() 返回的 gin 路由模式（含 ":id" 等占位符），避免任何 URL 改写绕过。
// 白名单仅覆盖仪表盘 + 使用记录的查询能力，所有写操作（创建/取消清理任务、聚合回填等）
// 全部拒绝。
//
// 维护提示：在 routes/admin.go 中新增 dashboard 或 usage 查询接口时，需要同步在此处
// 追加路由模式（method + FullPath()），否则 viewer 将收到 403 且不会有失败测试。
var viewerAllowed = map[string]struct{}{
	// 仪表盘
	"GET /api/v1/admin/dashboard/snapshot-v2":     {},
	"GET /api/v1/admin/dashboard/stats":           {},
	"GET /api/v1/admin/dashboard/realtime":        {},
	"GET /api/v1/admin/dashboard/trend":           {},
	"GET /api/v1/admin/dashboard/models":          {},
	"GET /api/v1/admin/dashboard/groups":          {},
	"GET /api/v1/admin/dashboard/api-keys-trend":  {},
	"GET /api/v1/admin/dashboard/users-trend":     {},
	"GET /api/v1/admin/dashboard/users-ranking":   {},
	"GET /api/v1/admin/dashboard/user-breakdown":  {},
	"POST /api/v1/admin/dashboard/users-usage":    {}, // body 携带查询参数，非写操作
	"POST /api/v1/admin/dashboard/api-keys-usage": {}, // body 携带查询参数，非写操作

	// 使用记录
	"GET /api/v1/admin/usage":                  {},
	"GET /api/v1/admin/usage/stats":            {},
	"GET /api/v1/admin/usage/search-users":     {},
	"GET /api/v1/admin/usage/search-api-keys":  {},
	"GET /api/v1/admin/usage/cleanup-tasks":    {},
}

// ViewerReadOnly 限制只读管理员（role=viewer）只能访问 viewerAllowed 白名单中的接口。
// 非 viewer 角色（admin 等）直接放行；viewer 命中白名单放行，未命中返回 403。
//
// 必须挂在 AdminAuthMiddleware 之后（依赖 ContextKeyUserRole）。
func ViewerReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRoleFromContext(c)
		if !ok {
			c.Next()
			return
		}
		if role != service.RoleViewer {
			c.Next()
			return
		}

		key := c.Request.Method + " " + c.FullPath()
		if _, allowed := viewerAllowed[key]; !allowed {
			AbortWithError(c, 403, "VIEWER_FORBIDDEN", "Viewer role is read-only; this endpoint is not permitted")
			return
		}
		c.Next()
	}
}
