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
	return func(c *gin.Context) {
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
