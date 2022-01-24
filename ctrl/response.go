package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/puras/mog/errdef"
	"github.com/puras/mog/response"
	"github.com/sirupsen/logrus"
)

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-24 12:09
 * @desc
 */

func RespError(c *gin.Context, err error) {
	if e, ok := err.(errdef.Error); ok {
		logrus.Info("yes")
		response.RespErr(c, e)
	} else {
		logrus.Info("no")
		response.RespFail(c, errdef.ServerException.Code, err.Error())
	}
}
