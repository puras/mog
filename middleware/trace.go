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

// 默认 header 键。提成常量供 DefaultTraceConfig 与 TraceWithConfig 的空值
// 归一共同引用，避免同一字面量在两处各写一遍后漂移。
//
// 不直接引用 DefaultTraceConfig：它是导出 var，运行期可被外部改写，
// 拿它当默认值会让归一结果随副作用变化。（Go 的 const 不支持 struct，
// 所以只能在字段级做常量。）
const (
	defaultTraceRequestHeaderKey = "X-Request-Id"
	defaultTraceResponseTraceKey = "X-Trace-Id"
)

var DefaultTraceConfig = TraceConfig{
	RequestHeaderKey: defaultTraceRequestHeaderKey,
	ResponseTraceKey: defaultTraceResponseTraceKey,
}

func Trace() gin.HandlerFunc {
	return TraceWithConfig(DefaultTraceConfig)
}

func TraceWithConfig(config TraceConfig) gin.HandlerFunc {
	// 空键归一，避免 Header().Set("") 写出无名响应头。
	if config.RequestHeaderKey == "" {
		config.RequestHeaderKey = defaultTraceRequestHeaderKey
	}
	if config.ResponseTraceKey == "" {
		config.ResponseTraceKey = defaultTraceResponseTraceKey
	}

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
		c.Request = c.Request.WithContext(ctx)

		// 入站读取与出站回写用不同的键。原先两处都用 RequestHeaderKey，
		// 导致 ResponseTraceKey 从未生效（配了也不写，永远是死配置）。
		//
		// 回写只走 ResponseTraceKey：请求头里的 id 可能由客户端带入，
		// 响应侧应当把服务端认定的 trace id 放在独立的键上，
		// 不必再把客户端原值回显到同一个头。
		c.Writer.Header().Set(config.ResponseTraceKey, traceId)

		c.Next()
	}
}
