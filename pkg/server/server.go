package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/**
* @project kuko
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-08-18 20:25
 */
type Server struct {
	ServerID string
	Address  string
}

func (s Server) String() string {
	return fmt.Sprintf("ServerID: %s, Address: %s", s.ServerID, s.Address)
}

func New() *Server {
	ID := fmt.Sprintf("%s-%s-%d", viper.GetString("registry.registry_name"), strings.Replace(viper.GetString("registry.register_address"), ".", "-", -1), viper.GetInt("registry.health_check_port"))
	return &Server{
		ServerID: ID,
		Address:  viper.GetString("server.addr"),
	}
}

func (s Server) Start(registryRouteFunc func(r *gin.Engine)) {
	r := InitRouter(registryRouteFunc)
	logrus.Infof("Start to listening on address: %s", s.Address)
	logrus.Info(http.ListenAndServe(s.Address, r).Error())
}
