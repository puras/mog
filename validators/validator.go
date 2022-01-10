package validators

import (
	"github.com/puras/mog/ctype"
	"reflect"
)

/**
 * @project kudo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-26 12:42
 * @desc
 */

func ValidateJSONDateType(field reflect.Value) interface{} {
	if field.Type() == reflect.TypeOf(ctype.Time{}) {
		timeStr := field.Interface().(ctype.Time).String()
		if timeStr == "0001-01-01 00:00:00" {
			return nil
		}
		return timeStr
	}
	return nil
}
