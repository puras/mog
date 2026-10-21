package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/web"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// withObservedLogger 把全局 logger 换成可观测的实例，返回日志缓冲与还原函数。
func withObservedLogger(t *testing.T) *observer.ObservedLogs {
	t.Helper()

	core, logs := observer.New(zapcore.DebugLevel)
	old := zap.L()
	zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(func() { zap.ReplaceGlobals(old) })

	return logs
}

// logsFor 取指定 msg 的日志条目。
func logsFor(logs *observer.ObservedLogs, msg string) []observer.LoggedEntry {
	var out []observer.LoggedEntry
	for _, e := range logs.All() {
		if e.Message == msg {
			out = append(out, e)
		}
	}
	return out
}

// 零值配置的归一结果必须与 DefaultLoggerConfig 一致。
// 修复前 LoggerWithConfig(LoggerConfig{}) 的请求体阈值是 4096，
// 而 DefaultLoggerConfig 是 1MB，同一配置项两种默认值。
func TestLogger_ZeroConfigMatchesDefault(t *testing.T) {
	if got, want := defaultMaxOutputReqBodyLen, DefaultLoggerConfig.MaxOutputRequestBodyLen; got != want {
		t.Errorf("请求体默认阈值不一致：归一 %d，DefaultLoggerConfig %d", got, want)
	}
	if got, want := defaultMaxOutputResBodyLen, DefaultLoggerConfig.MaxOutputResponseBodyLen; got != want {
		t.Errorf("响应体默认阈值不一致：归一 %d，DefaultLoggerConfig %d", got, want)
	}
	if got, want := DefaultLoggerConfig.RequestHeaderKey, defaultLogRequestHeaderKey; got != want {
		t.Errorf("header key 不一致：%q vs %q", got, want)
	}
}

// 大 body（超过旧的 4096 归一值、但小于 1MB）在零值配置下应当被完整记录。
func TestLogger_ZeroConfigRecordsLargeBody(t *testing.T) {
	logs := withObservedLogger(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerWithConfig(LoggerConfig{})) // 全零 → 走归一
	r.POST("/x", func(c *gin.Context) {
		body := make([]byte, 8192) // > 4096，< 1MB
		for i := range body {
			body[i] = 'a'
		}
		c.Set(web.RequestBodyKey, body)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/x", nil))

	entries := logsFor(logs, "[HTTP]")
	if len(entries) != 1 {
		t.Fatalf("访问日志条数 = %d, want 1", len(entries))
	}
	if body, ok := entries[0].ContextMap()["body"].(string); !ok || len(body) != 8192 {
		t.Errorf("8192 字节的请求体应被记录（阈值 1MB），实际 %v", entries[0].ContextMap()["body"])
	}
}

// 超过阈值的 body 不记录，避免日志爆量。
func TestLogger_SkipsBodyOverLimit(t *testing.T) {
	logs := withObservedLogger(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerWithConfig(LoggerConfig{MaxOutputRequestBodyLen: 8}))
	r.POST("/x", func(c *gin.Context) {
		c.Set(web.RequestBodyKey, []byte("0123456789")) // 10 > 8
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/x", nil))

	entries := logsFor(logs, "[HTTP]")
	if len(entries) != 1 {
		t.Fatalf("访问日志条数 = %d, want 1", len(entries))
	}
	if _, ok := entries[0].ContextMap()["body"]; ok {
		t.Error("超阈值 body 不应被记录")
	}
}

// request_id 从 header 抓取后应同时进 ctx 与响应/日志。
func TestLogger_InjectsTraceIdFromHeader(t *testing.T) {
	logs := withObservedLogger(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerWithConfig(LoggerConfig{}))
	r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-Id", "trace-from-client")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entries := logsFor(logs, "[HTTP]")
	if len(entries) != 1 {
		t.Fatalf("访问日志条数 = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["trace_id"]; got != "trace-from-client" {
		t.Errorf("trace_id = %v, want trace-from-client", got)
	}
}

// 跳过前缀不打印访问日志。
func TestLogger_SkippedPath(t *testing.T) {
	logs := withObservedLogger(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerWithConfig(LoggerConfig{SkippedPathPrefixes: []string{"/health"}}))
	r.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if n := len(logsFor(logs, "[HTTP]")); n != 0 {
		t.Errorf("跳过路径不应打印访问日志，实际 %d 条", n)
	}
	_ = strings.TrimSpace
}
