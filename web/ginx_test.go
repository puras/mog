package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/errors"
)

// ResError 会把非 *errors.Error 的错误交给 errors.InternalServerError —— 一个
// printf 变参函数。历史上这里是 InternalServerError("", err.Error())，等于把
// 运行时消息当作 format 字符串，消息里的每个 % 都被 fmt 解释成动词并渲染成
// %!s(MISSING)，损坏的文本既进日志又返回给客户端。
func TestResError_PreservesPercentInMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 含真实百分号的错误：url 解析失败会把用户的 percent 编码原样回显
	_, urlErr := url.Parse("http://host/api%zzpath")

	messages := []struct {
		label string
		err   error
	}{
		{"net/url 回显 percent 编码", urlErr},
		{"驱动回显 SQL 通配符", errString(`Error 1064: syntax near '%s'`)},
		{"含 %! 的消息", errString("broken: %!(NOVERB)")},
		{"裸百分号", errString("progress 100% of 3 tables")},
		{"动词带参数名", errString("cache hit ratio 95%s")},
		{"多动词无参数", errString("id %s name %d")},
		{"百分号在末尾", errString("trailing %")},
	}

	for _, tt := range messages {
		t.Run(tt.label, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

			ResError(c, tt.err)

			// 全等断言：一旦错误消息被当成 format 走过一次 Sprintf，
			// Message 就会变成 %!s(MISSING) 之类，与原文不再相等。
			var got ResponseResult
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("响应不是合法 JSON: %v (%s)", err, w.Body.String())
			}
			if got.Message != tt.err.Error() {
				t.Errorf("Message = %q\nwant     %q", got.Message, tt.err.Error())
			}
			if got.Code != "500" {
				t.Errorf("Code = %q, want %q", got.Code, "500")
			}
		})
	}
}

// 已经是 *errors.Error 的走另一条分支，同样不应被二次格式化。
func TestResError_TypedErrorPassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	ResError(c, errors.BadRequest("", "bad input: %s %d", "%", 7))

	var got ResponseResult
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if got.Message != "bad input: % 7" {
		t.Fatalf("类型化错误的 Detail 被改动: %q", got.Message)
	}
	if got.Code != "400" {
		t.Fatalf("Code = %q, want %q", got.Code, "400")
	}
}

// 显式传入 status 覆盖错误自带 Code 的路径。
func TestResError_StatusOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	ResError(c, errString("plain"), http.StatusUnauthorized)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("状态码 = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
