package validators

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
	"github.com/puras/mog/ctype"
)

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-01 14:24
 * @desc
 */
func InitCustomValid() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterCustomTypeFunc(ValidateJSONDateType, ctype.Time{})
	}
}
