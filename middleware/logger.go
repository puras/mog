package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/web"
	"go.uber.org/zap"
)

// LoggerConfig 是 HTTP 访问日志中间件的可调参数。
type LoggerConfig struct {
	SkippedPathPrefixes      []string
	AllowedPathPrefixes      []string
	MaxOutputRequestBodyLen  int
	MaxOutputResponseBodyLen int
	RequestHeaderKey         string // 默认 X-Request-Id
}

// DefaultLoggerConfig 提供基础默认值。
var DefaultLoggerConfig = LoggerConfig{
	MaxOutputRequestBodyLen:  1024 * 1024,
	MaxOutputResponseBodyLen: 1024 * 1024,
	RequestHeaderKey:         "X-Request-Id",
}

// Logger 装载默认配置 HTTP 访问日志中间件。
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig 通过 Config 装载 HTTP 访问日志中间件。
//
// 行为：
//   - 自动从 header 抓 trace_id 注入 ctx；
//   - 把 logger.TagKeyRequest 写入 tag；
//   - 结束统一用 logger.From(ctx).Info(...)，自动附带 trace_id/user_id/tag/span；
//   - 请求/响应 body 仅在 <配置长度阈值 时打印，规避大对象。
func LoggerWithConfig(cfg LoggerConfig) gin.HandlerFunc {
	if cfg.RequestHeaderKey == "" {
		cfg.RequestHeaderKey = "X-Request-Id"
	}
	if cfg.MaxOutputRequestBodyLen == 0 {
		cfg.MaxOutputRequestBodyLen = 4096
	}
	if cfg.MaxOutputResponseBodyLen == 0 {
		cfg.MaxOutputResponseBodyLen = 1024
	}

	return func(c *gin.Context) {
		if SkippedPathPrefixes(c, cfg.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// —— 注入 trace_id ——
		ctx := c.Request.Context()
		if v := c.GetHeader(cfg.RequestHeaderKey); v != "" {
			ctx = logger.NewTraceId(ctx, v)
			c.Request = c.Request.WithContext(ctx)
			c.Set("request_id", v)
		}
		ctx = logger.NewTag(ctx, logger.TagKeyRequest)

		start := time.Now()
		c.Next()

		cost := time.Since(start)
		fields := []zap.Field{
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("cost_ms", cost.Milliseconds()),
			zap.Int("res_size", c.Writer.Size()),
		}

		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if v, ok := c.Get(web.RequestBodyKey); ok {
				if b, ok := v.([]byte); ok && len(b) <= cfg.MaxOutputRequestBodyLen {
					fields = append(fields, zap.String("body", string(b)))
				}
			}
		}
		if v, ok := c.Get(web.ResponseBodyKey); ok {
			if b, ok := v.([]byte); ok && len(b) <= cfg.MaxOutputResponseBodyLen {
				fields = append(fields, zap.String("res_body", string(b)))
			}
		}

		logger.From(ctx).InfoF("[HTTP]", fields...)
	}
}
