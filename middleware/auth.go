package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/contextx"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/web"
)

// AuthInfo 描述一次认证解析出来的全部信息。
// Parser 把结果聚合到这里，再交给 Inject 写入 context；
// 后续增加字段（如 TenantId、Permissions 等）只需扩展此结构与 Inject，
// 不再影响 Parse 的签名和外部调用方。
type AuthInfo struct {
	UserId string
	Role   string
}

// Inject 把认证信息写入 context，便于下游通过 contextx 取用。
func (a *AuthInfo) Inject(ctx context.Context) context.Context {
	if a == nil {
		return ctx
	}
	ctx = contextx.NewUserId(ctx, a.UserId)
	ctx = contextx.NewRole(ctx, a.Role)
	return ctx
}

type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Skipper             func(*gin.Context) bool
	Parse               func(*gin.Context) (*AuthInfo, error)
}

func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	// Parse 缺失意味着认证逻辑不存在。运行期再调用会 nil panic（且发生在请求
	// 中途，返回 500 而不是 401）；静默放行则等于把认证关掉，风险更大。
	// 启动即失败是唯一安全的选择。
	if config.Parse == nil {
		panic("middleware.Auth: Parse 不能为 nil")
	}

	return func(c *gin.Context) {
		// CORS 预检请求不携带凭证,必须放行,由 CORS 中间件响应
		// (否则 Auth 先于 CORS 执行,预检被 401,浏览器报无 Access-Control-Allow-Origin)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		//if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		info, err := config.Parse(c)
		if err != nil {
			web.ResError(c, err)
			return
		}

		ctx := info.Inject(c.Request.Context())
		ctx = logger.NewUserId(ctx, info.UserId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
