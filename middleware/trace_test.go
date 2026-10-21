package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/logger"
)

// newTraceRouter 装一个只回 "ok" 的路由，用于观察响应头。
func newTraceRouter(cfg TraceConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TraceWithConfig(cfg))
	r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func doTraceRequest(r *gin.Engine, inbound map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	for k, v := range inbound {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

// 出站回写只走 ResponseTraceKey。修复前两处都用 RequestHeaderKey，
// X-Trace-Id 永远是空的（ResponseTraceKey 是死配置）。
func TestTrace_WritesResponseTraceKey(t *testing.T) {
	w := doTraceRequest(newTraceRouter(DefaultTraceConfig), nil)

	traceId := w.Header().Get("X-Trace-Id")
	if traceId == "" {
		t.Fatalf("X-Trace-Id 为空：ResponseTraceKey 未被使用")
	}
	if !strings.HasPrefix(traceId, "trace-") {
		t.Errorf("未生成 trace 前缀：%q", traceId)
	}

	// 回写键是 ResponseTraceKey，不再占用请求键
	if got := w.Header().Get("X-Request-Id"); got != "" {
		t.Errorf("不该再回写 X-Request-Id，得到 %q", got)
	}
}

// 客户端自带 trace id 时沿用（入站读 RequestHeaderKey），
// 但只通过 ResponseTraceKey 回传。
func TestTrace_EchoesClientTraceId(t *testing.T) {
	w := doTraceRequest(newTraceRouter(DefaultTraceConfig), map[string]string{
		"X-Request-Id": "client-supplied-123",
	})

	if got := w.Header().Get("X-Trace-Id"); got != "client-supplied-123" {
		t.Errorf("X-Trace-Id = %q, want %q", got, "client-supplied-123")
	}
	if got := w.Header().Get("X-Request-Id"); got != "" {
		t.Errorf("不该 echoed 回 X-Request-Id，得到 %q", got)
	}
}

// 自定义键：入站读 X-In，出站写 X-Out。这是修复前的失败点——X-Out 恒为空。
func TestTrace_CustomKeys(t *testing.T) {
	w := doTraceRequest(newTraceRouter(TraceConfig{
		RequestHeaderKey: "X-In",
		ResponseTraceKey: "X-Out",
	}), map[string]string{
		"X-In": "abc",
	})

	if got := w.Header().Get("X-Out"); got != "abc" {
		t.Errorf("X-Out = %q, want %q（ResponseTraceKey 未生效）", got, "abc")
	}
	if got := w.Header().Get("X-In"); got != "" {
		t.Errorf("不该回写请求键 X-In，得到 %q", got)
	}
}

// 两键相同时只写一次，不产生重复 header。
func TestTrace_SameKeys(t *testing.T) {
	w := doTraceRequest(newTraceRouter(TraceConfig{
		RequestHeaderKey: "X-Trace",
		ResponseTraceKey: "X-Trace",
	}), nil)

	values := w.Header().Values("X-Trace")
	if len(values) != 1 {
		t.Fatalf("X-Trace 出现 %d 次，want 1: %v", len(values), values)
	}
	if values[0] == "" {
		t.Errorf("X-Trace 值为空")
	}
}

// 零值配置必须归一为默认键，不能写出无名响应头。
func TestTrace_EmptyConfig(t *testing.T) {
	w := doTraceRequest(newTraceRouter(TraceConfig{}), nil)

	if got := w.Header().Get(defaultTraceResponseTraceKey); got == "" {
		t.Errorf("%s 为空：空配置未归一", defaultTraceResponseTraceKey)
	}
	// Header().Set("") 会写出一个 key 为空的畸形头
	if _, ok := w.Header()[""]; ok {
		t.Errorf("产生了无名响应头：%v", w.Header())
	}
}

// 命中跳过前缀时不应写任何追踪响应头。
func TestTrace_SkippedPath(t *testing.T) {
	w := doTraceRequest(newTraceRouter(TraceConfig{
		SkippedPathPrefixes: []string{"/x"},
		RequestHeaderKey:    defaultTraceRequestHeaderKey,
		ResponseTraceKey:    defaultTraceResponseTraceKey,
	}), nil)

	if got := w.Header().Get(defaultTraceRequestHeaderKey); got != "" {
		t.Errorf("跳过路径不该写 %s，得到 %q", defaultTraceRequestHeaderKey, got)
	}
	if got := w.Header().Get(defaultTraceResponseTraceKey); got != "" {
		t.Errorf("跳过路径不该写 %s，得到 %q", defaultTraceResponseTraceKey, got)
	}
}

// 响应头里的 trace id 必须与注入 ctx 的一致——日志按它检索才查得到。
func TestTrace_InjectsLoggerCtx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TraceWithConfig(DefaultTraceConfig))

	var fromCtx string
	r.GET("/x", func(c *gin.Context) {
		fromCtx = logger.FromTraceId(c.Request.Context())
		c.String(http.StatusOK, "ok")
	})

	w := doTraceRequest(r, nil)
	if fromCtx == "" {
		t.Fatal("ctx 里没有 trace id")
	}
	if fromCtx != w.Header().Get(defaultTraceResponseTraceKey) {
		t.Errorf("ctx trace id = %q，响应头 %s = %q，二者不一致",
			fromCtx, defaultTraceResponseTraceKey, w.Header().Get(defaultTraceResponseTraceKey))
	}
}
