package mog

import (
	"net/http"

	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 21:21
 */
func InitRouter(registryRouteFunc func(r *gin.Engine)) *gin.Engine {
	var router *gin.Engine

	runMode := viper.GetString("runmode")
	if runMode == gin.ReleaseMode {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(runMode)
	}
	router = gin.Default()
	router.Use(gin.Recovery())
	router.Use(Logging())

	InitCustomValid()

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

	//if runMode == gin.DebugMode {
	//	registrySwaggerRouter(router)
	//}

	registryRouteFunc(router)

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route.")
	})
	return router
}
