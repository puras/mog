package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/puras/mog/contextx"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/web"
)

type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Skipper             func(*gin.Context) bool
	ParseUser           func(*gin.Context) (string, error)
}

func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		//if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		userId, err := config.ParseUser(c)
		if err != nil {
			web.ResError(c, err)
			return
		}

		ctx := contextx.NewUserId(c.Request.Context(), userId)
		ctx = logger.NewUserId(ctx, userId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
