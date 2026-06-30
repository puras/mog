package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/inject"
	"github.com/puras/mog/middleware"
)

type ServerParam interface {
	Init(ctx context.Context) error
	GetInjector(ctx context.Context) *inject.Injector
	RegistryRoutes(ctx context.Context, e *gin.Engine) error
	ParseCurrentUser(c *gin.Context) (*middleware.AuthInfo, error)
}
