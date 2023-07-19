package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/puras/mog/inject"
)

type ServerParam interface {
	Init(ctx context.Context) error
	GetInjector(ctx context.Context) *inject.Injector
	RegistryRoutes(ctx context.Context, e *gin.Engine) error
	ParseCurrentUser(c *gin.Context) (string, error)
}
