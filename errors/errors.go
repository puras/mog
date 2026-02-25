package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

var (
	WithStack = errors.WithStack
	Wrap      = errors.Wrap
	Wrapf     = errors.Wrapf
)

const (
	DefaultBadRequestId            = "bad_request"
	DefaultUnauthorizedId          = "unauthorized"             // 未授权
	DefaultForbiddenId             = "forbidden"                // 无操作权限
	DefaultNotFoundId              = "not_found"                // 数据未找到
	DefaultBindError               = "bind_error"               // 绑定错误
	DefaultDataIsExists            = "data_is_exists"           // 数据已存在
	DefaultDataNotAllowEdit        = "data_not_allow_edit"      // 数据不允许编辑
	DefaultDataCheckFailure        = "data_check_failure"       // 数据检查失败
	DefaultDataIsRelation          = "data_is_relation"         // 数据被引用
	DefaultDataParseFailure        = "data_parse_failure"       // 数据解析失败
	DefaultBizError                = "data_biz_error"           //业务逻辑错误
	DefaultInvalidParam            = "invalid_param"            // 无效的参数
	DefaultInvalidJson             = "invalid_json"             // 无效的JSON串
	DefaultInvalidToken            = "invalid_token"            // 无效的Token
	DefaultRemoteCallError         = "remote_call_error"        //远程调用错误
	DefaultMethodNotAllowedId      = "method_not_allowed"       // 方法不支持
	DefaultTooManyRequestsId       = "too_many_requests"        // 太多的请求
	DefaultRequestEntityTooLargeId = "request_entity_too_large" // 请求体过大
	DefaultInternalServerErrorId   = "internal_server_error"    // 服务端错误
	DefaultConflictId              = "conflict"                 // 冲突
	DefaultRequestTimeoutId        = "request_timeout"          // 请求超时
)

type Error struct {
	Id     string `json:"id,omitempty"`
	Code   int32  `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
	Status string `json:"status,omitempty"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func New(id, detail string, code int32) error {
	return &Error{
		Id:     id,
		Code:   code,
		Detail: detail,
		Status: http.StatusText(int(code)),
	}
}

func Parse(err string) *Error {
	e := new(Error)
	er := json.Unmarshal([]byte(err), e)
	if er != nil {
		e.Detail = err
	}
	return e
}
func BadRequest(id, format string, a ...any) error {
	if id == "" {
		id = DefaultBadRequestId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusBadRequest,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusBadRequest),
	}
}

func Unauthorized(id, format string, a ...any) error {
	if id == "" {
		id = DefaultUnauthorizedId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusUnauthorized,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusUnauthorized),
	}
}

func Forbidden(id, format string, a ...any) error {
	if id == "" {
		id = DefaultForbiddenId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusForbidden,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusForbidden),
	}
}

func NotFound(id, format string, a ...any) error {
	if id == "" {
		id = DefaultNotFoundId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusNotFound,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusNotFound),
	}
}

func BindError(id, format string, a ...any) error {
	if id == "" {
		id = DefaultBindError
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func DataIsExists(id, format string, a ...any) error {
	if id == "" {
		id = DefaultDataIsExists
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func DataNotAllowEdit(id, format string, a ...any) error {
	if id == "" {
		id = DefaultDataNotAllowEdit
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func DataCheckFailure(id, format string, a ...any) error {
	if id == "" {
		id = DefaultDataCheckFailure
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func DataIsRelation(id, format string, a ...any) error {
	if id == "" {
		id = DefaultDataIsRelation
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func DataParseFailure(id, format string, a ...any) error {
	if id == "" {
		id = DefaultDataParseFailure
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func BizError(id, format string, a ...any) error {
	if id == "" {
		id = DefaultBizError
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func InvalidParam(id, format string, a ...any) error {
	if id == "" {
		id = DefaultInvalidParam
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func InvalidJson(id, format string, a ...any) error {
	if id == "" {
		id = DefaultInvalidJson
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func InvalidToken(id, format string, a ...any) error {
	if id == "" {
		id = DefaultInvalidToken
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func RemoteCallError(id, format string, a ...any) error {
	if id == "" {
		id = DefaultRemoteCallError
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func MethodNotAllowed(id, format string, a ...any) error {
	if id == "" {
		id = DefaultMethodNotAllowedId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusMethodNotAllowed,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusMethodNotAllowed),
	}
}

func TooManyRequests(id, format string, a ...any) error {
	if id == "" {
		id = DefaultTooManyRequestsId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusTooManyRequests,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusTooManyRequests),
	}
}

func Timeout(id, format string, a ...any) error {
	if id == "" {
		id = DefaultRequestTimeoutId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusRequestTimeout,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestTimeout),
	}
}

func Conflict(id, format string, a ...any) error {
	if id == "" {
		id = DefaultConflictId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusConflict,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusConflict),
	}
}

func RequestEntityTooLarge(id, format string, a ...any) error {
	if id == "" {
		id = DefaultRequestEntityTooLargeId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusRequestEntityTooLarge,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestEntityTooLarge),
	}
}

func InternalServerError(id, format string, a ...any) error {
	if id == "" {
		id = DefaultInternalServerErrorId
	}
	return &Error{
		Id:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}
	if !ok1 {
		return err1 == err2
	}
	if verr1.Code != verr2.Code {
		return false
	}
	return true
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*Error); ok && verr != nil {
		return verr
	}
	return Parse(err.Error())
}

func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
