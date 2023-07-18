package middleware

import (
	"fmt"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/web"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecoveryConfig struct {
	Skip int
}

var DefaultRecoveryConfig = RecoveryConfig{
	Skip: 3,
}

func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rv := recover(); rv != nil {
				ctx := c.Request.Context()
				ctx = logger.NewTag(ctx, logger.TagKeyRecovery)

				var fields []zap.Field
				fields = append(fields, zap.StackSkip("stack", config.Skip))

				if gin.IsDebugging() {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					fields = append(fields, zap.Strings("headers", headers))
				}

				logger.Context(ctx).Error(fmt.Sprintf("[Recovery] %s panic recovered", time.Now().Format("2006-01-02 15:04:05")), fields...)
				web.ResError(c, errors.InternalServerError("", "Internal server error, please try again later"))
			}
		}()

		c.Next()
	}
}
