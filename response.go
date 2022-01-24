package mog

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 11:16
 * @desc
 */

type Response struct {
	Code      string      `json:"code"`
	RequestId string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
}

func (r *Response) SetData(data interface{}) {
	r.Data = data
}

func (r *Response) SetTimestamp(timestamp int64) {
	r.Timestamp = timestamp
}

func (r *Response) SetRequestId(requestId string) {
	r.RequestId = requestId
}

func GetRequestId(c *gin.Context) string {
	requestId := c.Request.Header.Get("X-Request-Id")
	if requestId == "" {
		requestId = GenShortUUID()
	}
	return requestId
}

func RespOk(c *gin.Context, data interface{}) {
	resp := Response{Code: SUCCESS, RequestId: GetRequestId(c), Data: data, Timestamp: time.Now().Unix()}
	c.JSON(http.StatusOK, resp)
}

func RespErrCode(c *gin.Context, err ErrCode) {
	resp := Response{RequestId: GetRequestId(c), Code: err.Code, Message: err.Message}
	c.JSON(http.StatusBadRequest, resp)
}

func RespErr(c *gin.Context, err Error) {
	resp := Response{RequestId: GetRequestId(c), Code: err.Code, Message: err.Message}
	c.JSON(http.StatusBadRequest, resp)
}

func RespFail(c *gin.Context, code string, message string) {
	resp := Response{Code: code, RequestId: GetRequestId(c), Message: message}
	c.JSON(http.StatusBadRequest, resp)
}

func RespError(c *gin.Context, err error) {
	if e, ok := err.(Error); ok {
		logrus.Info("yes")
		RespErr(c, e)
	} else {
		logrus.Info("no")
		RespFail(c, ErrServerException.Code, err.Error())
	}
}
