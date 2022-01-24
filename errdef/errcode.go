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

var Unauthorized = ErrCode{"Unauthorized", "未经授权"}
var Forbidden = ErrCode{"Forbidden", "无操作权限"}
var DataNotFound = ErrCode{"DataNotFound", "数据未找到"}
var RelationNotFound = ErrCode{"RelationNotFound", "关联数据未找到"}
var ServerException = ErrCode{"ServerException", "服务端异常"}
var BindError = ErrCode{"BindError", "数据绑定错误"} // 数据类96开头
var DataIsExist = ErrCode{"DataIsExist", "同类数据已存在"}
var DataSelfReference = ErrCode{"DataSelfReference", "数据进行自引用"}
var DataNotEditable = ErrCode{"DataNotEditable", "数据不可编辑"}
var DataCheckFailure = ErrCode{"DataCheckFailure", "数据检查失败"}
var DataIsRelation = ErrCode{"DataIsRelation", "数据被引用"}
var DataParseFailure = ErrCode{"DataParseFailure", "数据解析失败"}
var DataOperateFailure = ErrCode{"DataOperateFailure", "数据操作失败"}
var DataEncodeFailure = ErrCode{"DataEncodeFailure", "数据编码失败"}
var DataDecodeFailure = ErrCode{"DataDecodeFailure", "数据解码失败"}
var BusinessError = ErrCode{"BusinessError", "业务逻辑错误"}
var InternalError = ErrCode{"InternalError", "服务端异常"}
var InvalidParam = ErrCode{"InvalidParameter", "无效的请求参数"}
var InvalidJson = ErrCode{"InvalidJson", "无效的JSON请求串"}
var InvalidToken = ErrCode{"InvalidToken", "无效的令牌"}
var AuthFailed = ErrCode{"AuthFailed", "鉴权失败"}
var ExtNameMismatch = ErrCode{"ExtNameMismatch", "扩展名不匹配"}
var FileSizeTooLarge = ErrCode{"FileSizeTooLarge", "文件尺寸过大"}
var RemoteCallError = ErrCode{"RemoteCallError", "远程调用失败"}
var RequestMethodNotSupported = ErrCode{"RequestMethodNotSupported", "请求的HTTP方法不支持"}
var MethodNotSupported = ErrCode{"MethodNotSupported", "方法不支持"}
var NoSolution = ErrCode{"NoSolution", "无解了"}
