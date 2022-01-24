package mog

import (
	"reflect"
)

/**
 * @project kudo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-26 12:42
 * @desc
 */

func ValidateJSONDateType(field reflect.Value) interface{} {
	if field.Type() == reflect.TypeOf(Time{}) {
		timeStr := field.Interface().(Time).String()
		if timeStr == "0001-01-01 00:00:00" {
			return nil
		}
		return timeStr
	}
	return nil
}
