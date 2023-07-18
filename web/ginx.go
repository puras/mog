package web

import (
	"encoding/json"
	"github.com/puras/mog/dbx"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

const (
	RequestBodyKey  = "request-body"
	ResponseBodyKey = "response-body"

	Success = "0"
)

type ResponseResult struct {
	Code    string `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type PageResult struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
}

func FromPaginationResult(pr *dbx.PaginationResult) *PageResult {
	return &PageResult{
		Items: pr.Items,
		Total: pr.Total,
	}
}

func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(RequestBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

func ParseJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.BadRequest("", "Failed to parse json: %s", err.Error())
	}
	return nil
}

func ParseQuery(c *gin.Context, obj any) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.BadRequest("", "Failed to parse query: %s", err.Error())
	}
	return nil
}

func ParseForm(c *gin.Context, obj any) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.BadRequest("", "Failed to parse form: %s", err.Error())
	}
	return nil
}

func ResJson(c *gin.Context, status int, v any) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	c.Set(ResponseBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

func ResSuccess(c *gin.Context, v any) {
	ResJson(c, http.StatusOK, ResponseResult{
		Code: Success,
		Data: v,
	})
}

func ResOk(c *gin.Context) {
	ResJson(c, http.StatusOK, NewResponseResult(Success, "", nil))
}

func ResPage(c *gin.Context, pr *PageResult) {
	ResSuccess(c, pr)
}

//func ResPagination(c *gin.Context, v any, total int64) {
//	reflectValue := reflect.Indirect(reflect.ValueOf(v))
//	if reflectValue.IsNil() {
//		v = make([]any, 0)
//	}
//	ResSuccess(c, PageResult{
//		Items: v,
//		Total: total,
//	})
//}

func ResError(c *gin.Context, err error, status ...int) {
	ctx := c.Request.Context()
	var er *errors.Error
	if e, ok := errors.As(err); ok {
		er = e
	} else {
		er = errors.FromError(errors.InternalServerError("", err.Error()))
	}

	code := int(er.Code)
	if len(status) > 0 {
		code = status[0]
	}

	fields := []zap.Field{
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("referer", c.Request.Referer()),
		zap.String("uri", c.Request.RequestURI),
		zap.String("host", c.Request.Host),
		zap.String("remote_addr", c.Request.RemoteAddr),
		zap.String("proto", c.Request.Proto),
		zap.Int64("content_length", c.Request.ContentLength),
		zap.String("pragma", c.Request.Header.Get("Pragma")),
		zap.Int("code", code),
		zap.Error(err),
	}

	ctx = logger.NewTag(ctx, logger.TagKeySystem)
	if code >= 400 && code < 500 {
		logger.Context(ctx).Info(er.Detail, fields...)
	} else if code >= 500 {
		logger.Context(ctx).Error(er.Detail, fields...)
	}

	if code >= 500 {
		er.Detail = http.StatusText(http.StatusInternalServerError)
	}
	er.Code = int32(code)
	ResJson(c, code, NewResponseResult(strconv.Itoa(code), er.Detail, nil))
}

func NewResponseResult(code string, message string, data any) ResponseResult {
	return ResponseResult{
		Code:    code,
		Data:    data,
		Message: message,
	}
}
