package mog

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

var ErrUnauthorized = ErrCode{"Unauthorized", "未经授权"}
var ErrForbidden = ErrCode{"Forbidden", "无操作权限"}
var ErrDataNotFound = ErrCode{"DataNotFound", "数据未找到"}
var ErrRelationNotFound = ErrCode{"RelationNotFound", "关联数据未找到"}
var ErrServerException = ErrCode{"ServerException", "服务端异常"}
var ErrBindError = ErrCode{"BindError", "数据绑定错误"} // 数据类96开头
var ErrDataIsExist = ErrCode{"DataIsExist", "同类数据已存在"}
var ErrDataSelfReference = ErrCode{"DataSelfReference", "数据进行自引用"}
var ErrDataNotEditable = ErrCode{"DataNotEditable", "数据不可编辑"}
var ErrDataCheckFailure = ErrCode{"DataCheckFailure", "数据检查失败"}
var ErrDataIsRelation = ErrCode{"DataIsRelation", "数据被引用"}
var ErrDataParseFailure = ErrCode{"DataParseFailure", "数据解析失败"}
var ErrDataOperateFailure = ErrCode{"DataOperateFailure", "数据操作失败"}
var ErrDataEncodeFailure = ErrCode{"DataEncodeFailure", "数据编码失败"}
var ErrDataDecodeFailure = ErrCode{"DataDecodeFailure", "数据解码失败"}
var ErrBusinessError = ErrCode{"BusinessError", "业务逻辑错误"}
var ErrInternalError = ErrCode{"InternalError", "服务端异常"}
var ErrInvalidParam = ErrCode{"InvalidParameter", "无效的请求参数"}
var ErrInvalidJson = ErrCode{"InvalidJson", "无效的JSON请求串"}
var ErrInvalidToken = ErrCode{"InvalidToken", "无效的令牌"}
var ErrAuthFailed = ErrCode{"AuthFailed", "鉴权失败"}
var ErrExtNameMismatch = ErrCode{"ExtNameMismatch", "扩展名不匹配"}
var ErrFileSizeTooLarge = ErrCode{"FileSizeTooLarge", "文件尺寸过大"}
var ErrRemoteCallError = ErrCode{"RemoteCallError", "远程调用失败"}
var ErrRequestMethodNotSupported = ErrCode{"RequestMethodNotSupported", "请求的HTTP方法不支持"}
var ErrMethodNotSupported = ErrCode{"MethodNotSupported", "方法不支持"}
var ErrNoSolution = ErrCode{"NoSolution", "无解了"}
