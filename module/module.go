package module

import (
	"context"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Module interface {
	Init(context.Context) error
	RegistryRoutes(context.Context, *gin.Engine) error
}

func Init(ctx context.Context, fields reflect.Value) error {
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		ret, ok := field.Interface().(Module)
		if ok {
			if err := ret.Init(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func RegistryRoutes(ctx context.Context, e *gin.Engine, fields reflect.Value) error {
	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		ret, ok := field.Interface().(Module)
		if ok {
			if err := ret.RegistryRoutes(ctx, e); err != nil {
				return err
			}
		}
	}
	return nil
}
