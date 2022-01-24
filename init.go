package mog

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
)

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-01 14:24
 * @desc
 */
func InitCustomValid() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(ValidateJSONDateType, Time{})
	}
}
