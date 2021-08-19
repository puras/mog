package errcode

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 13:25
 * @desc
 */
type ErrCode struct {
	Code    string
	Message string
}

var UNAUTHORIZED = ErrCode{Code: "Unauthorized", Message: "未经授权"}
var FORBIDDEN = ErrCode{Code: "Forbidden", Message: "无操作权限"}
