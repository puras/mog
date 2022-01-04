package ctrl

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"mooko.net/mog/pkg/errdef"
	"mooko.net/mog/pkg/response"
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
