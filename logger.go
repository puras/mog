package mog

import (
	"bytes"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

/**
 * @project kudo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-01 12:24
 * @desc
 */
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		ct := c.Request.Header.Get("Content-Type")
		isForm := strings.HasPrefix(ct, "multipart/form-data")

		var bodyBytes []byte
		if !isForm {
			if c.Request.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
			}

			if string(bodyBytes) == "" {
				bodyBytes = []byte("{}")
			}

			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}

		c.Writer = blw
		c.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqUri := c.Request.RequestURI
		reqMethod := c.Request.Method
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logrus.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
			"req_body":     string(bodyBytes),
		}).Info(blw.body)
	}
}
