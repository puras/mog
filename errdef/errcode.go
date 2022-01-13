package errdef

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

var UNAUTHORIZED = ErrCode{"Unauthorized", "未经授权"}
var FORBIDDEN = ErrCode{"Forbidden", "无操作权限"}
var DataNotFound = ErrCode{"DataNotFound", "数据未找到"}
var ServerException = ErrCode{"ServerException", "服务端异常"}
var BindError = ErrCode{"BindError", "数据绑定错误"} // 数据类96开头
var DataIsExist = ErrCode{"DataIsExist", "同类数据已存在"}
var DataSelfReference = ErrCode{"DataSelfReference", "数据进行自引用"}
var DATA_NOT_EDITABLE = ErrCode{"DataNotEditable", "数据不可编辑"}
var DATA_CHECK_FAILURE = ErrCode{"DataCheckFailure", "数据检查失败"}
var DATA_IS_RELATION = ErrCode{"DataIsRelation", "数据被引用"}
var DATA_PARSE_FAILURE = ErrCode{"DataParseFailure", "数据解析失败"}
var DATA_ENCODE_FAILURE = ErrCode{"DataEncodeFailure", "数据编码失败"}
var DATA_DECODE_FAILURE = ErrCode{"DataDecodeFailure", "数据解码失败"}
var BUSINESS_ERROR = ErrCode{"BusinessError", "业务逻辑错误"}
var INTERNAL_ERROR = ErrCode{"InternalError", "服务端异常"}
var INVALID_PARAM = ErrCode{"InvalidParameter", "无效的请求参数"}
var INVALID_JSON = ErrCode{"InvalidJson", "无效的JSON请求串"}
var INVALID_TOKEN = ErrCode{"InvalidToken", "无效的令牌"}
var AUTH_FAILED = ErrCode{"AuthFailed", "鉴权失败"}
var EXT_NAME_MISMATCH = ErrCode{"ExtNameMismatch", "扩展名不匹配"}
var FileTypeMisMatch = ErrCode{"FileTypeMisMatch", "文件类型不匹配"}
var FileSizeToLarge = ErrCode{"FileSizeToLarge", "文件尺寸过大"}
var REMOTE_CALL_ERROR = ErrCode{"RemoteCallError", "远程调用失败"}
var REQUEST_METHOD_NOT_SUPPORTED = ErrCode{"RequestMethodNotSupported", "请求的HTTP方法不支持"}
var NEVER_USED_CODE = ErrCode{"NoSolution", "无解了"}
