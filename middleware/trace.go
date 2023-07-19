package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/puras/mog/logger"
	"github.com/rs/xid"
)

type TraceConfig struct {
	SkippedPathPrefixes []string
	AllowedPathPrefixes []string
	RequestHeaderKey    string
	ResponseTraceKey    string
}

var DefaultTraceConfig = TraceConfig{
	RequestHeaderKey: "X-Request-Id",
	ResponseTraceKey: "X-Trace-Id",
}

func Trace() gin.HandlerFunc {
	return TraceWithConfig(DefaultTraceConfig)
}

func TraceWithConfig(config TraceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		traceId := c.GetHeader(config.RequestHeaderKey)
		if traceId == "" {
			traceId = fmt.Sprintf("trace-%s", xid.New().String())
		}

		ctx := logger.NewTraceId(c.Request.Context(), traceId)
		ctx = logger.NewTraceId(ctx, traceId)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(config.RequestHeaderKey, traceId)

		c.Next()
	}
}
