package middleware

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/web"

	"go.uber.org/zap"
)

// RecoveryConfig Recovery 中间件可调参数。
type RecoveryConfig struct {
	Skip int // caller skip
}

var DefaultRecoveryConfig = RecoveryConfig{Skip: 3}

// Recovery 安装默认配置 Recovery 中间件。
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// RecoveryWithConfig 安装可配置 Recovery 中间件，panic 时打印堆栈并返回 500。
func RecoveryWithConfig(cfg RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rv := recover(); rv != nil {
				ctx := logger.NewTag(c.Request.Context(), logger.TagKeyRecovery)

				fields := []zap.Field{
					zap.StackSkip("stack", cfg.Skip),
					zap.String("time", time.Now().Format("2006-01-02 15:04:05")),
					zap.Any("panic", rv),
				}

				if gin.IsDebugging() {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if len(current) > 0 && current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					fields = append(fields, zap.Strings("headers", headers))
				}

				logger.From(ctx).ErrorF("[Recovery] panic recovered", fields...)
				if !c.Writer.Written() {
					detail := fmt.Sprintf("%v", rv)
					web.ResError(c, errors.New("", detail, http.StatusInternalServerError))
				}
			}
		}()

		c.Next()
	}
}
