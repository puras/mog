package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/contextx"
	"github.com/puras/mog/errors"
)

// Parse 为 nil 时必须启动即 panic，而不是运行期崩或静默放行。
func TestAuthWithConfig_NilParsePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Parse 为 nil 时应 panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "Parse") {
			t.Fatalf("panic 信息应指明 Parse，得到 %v", r)
		}
	}()
	AuthWithConfig(AuthConfig{})
}

// 成功路径：AuthInfo 注入 ctx，下游可取到 user_id / role。
func TestAuthWithConfig_InjectsUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthWithConfig(AuthConfig{
		Parse: func(c *gin.Context) (*AuthInfo, error) {
			return &AuthInfo{UserId: "u1", Role: "admin"}, nil
		},
	}))

	var gotUser, gotRole string
	r.GET("/x", func(c *gin.Context) {
		gotUser = contextx.FromUserId(c.Request.Context())
		gotRole = contextx.FromRole(c.Request.Context())
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if gotUser != "u1" || gotRole != "admin" {
		t.Errorf("ctx 中 user_id=%q role=%q，want u1/admin", gotUser, gotRole)
	}
}

// Parse 返回错误时交给 web.ResError，且不进入 handler。
func TestAuthWithConfig_ParseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthWithConfig(AuthConfig{
		Parse: func(c *gin.Context) (*AuthInfo, error) {
			return nil, errors.Unauthorized("", "token invalid")
		},
	}))

	reached := false
	r.GET("/x", func(c *gin.Context) {
		reached = true
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if reached {
		t.Error("认证失败时不应进入 handler")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "token invalid") {
		t.Errorf("响应体未包含错误详情: %s", w.Body.String())
	}
}

// 命中 SkippedPathPrefixes 时不应调用 Parse。
func TestAuthWithConfig_Skipped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false
	r.Use(AuthWithConfig(AuthConfig{
		SkippedPathPrefixes: []string{"/open"},
		Parse: func(c *gin.Context) (*AuthInfo, error) {
			called = true
			return &AuthInfo{}, nil
		},
	}))
	r.GET("/open/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/open/x", nil))

	if called {
		t.Error("跳过路径仍调用了 Parse")
	}
	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want 200", w.Code)
	}
}

// CORS 预检请求不携带凭证，必须放行且不调 Parse。
func TestAuthWithConfig_OptionsBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false
	r.Use(AuthWithConfig(AuthConfig{
		Parse: func(c *gin.Context) (*AuthInfo, error) {
			called = true
			return nil, errors.Unauthorized("", "should not be called")
		},
	}))
	r.OPTIONS("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/x", nil))

	if called {
		t.Error("OPTIONS 预检不应调用 Parse")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("状态码 = %d, want 204", w.Code)
	}
}
