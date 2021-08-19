package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 21:21
 */
func InitRouter(registryRouteFunc func(r *gin.Engine)) *gin.Engine {
	router := gin.New()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "you are welcome",
		})
	})

	router.GET("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "health check",
		})
	})

	registryRouteFunc(router)

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route.")
	})
	return router
}
