package handler

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 22:37
 */

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"mooko.net/mog/pkg/errcode"
	"mooko.net/mog/pkg/response"
	"mooko.net/mog/service"
)

var srv = service.LibraryService{}

func List(c *gin.Context) {
	list, err := srv.FindBy()
	if err != nil {
		response.RespByErrCode(c, errcode.FORBIDDEN)
	}
	fmt.Println(list)
	response.RespOk(c, list)
}
