package mog

import (
	"fmt"
)

/**
 * @project kudo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-26 13:37
 * @desc
 */

// Error 错误
type Error struct {
	ErrCode
}

func NewError(err ErrCode) Error {
	return Error{err}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
