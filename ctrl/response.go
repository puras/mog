package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	errdef2 "mooko.net/mog/errdef"
	"mooko.net/mog/response"
)

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-24 12:09
 * @desc
 */

func RespError(c *gin.Context, err error) {
	if e, ok := err.(errdef2.Error); ok {
		logrus.Info("yes")
		response.RespErr(c, e)
	} else {
		logrus.Info("no")
		response.RespFail(c, errdef2.ServerException.Code, err.Error())
	}
}
