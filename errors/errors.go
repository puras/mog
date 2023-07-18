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
	DefaultUnauthorizedId          = "unauthorized"
	DefaultForbiddenId             = "forbidden"
	DefaultNotFoundId              = "not_found"
	DefaultMethodNotAllowedId      = "method_not_allowed"
	DefaultTooManyRequestsId       = "too_many_requests"
	DefaultRequestEntityTooLargeId = "request_entity_too_large"
	DefaultInternalServerErrorId   = "internal_server_error"
	DefaultConflictId              = "conflict"
	DefaultRequestTimeoutId        = "request_timeout"
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
