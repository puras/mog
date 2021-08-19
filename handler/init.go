package handler

import (
	"github.com/gin-gonic/gin"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 22:42
 */
func RegistryRoute(r *gin.Engine) {
	r.GET("/libraries", List)
}
