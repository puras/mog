package server

import (
	"context"
	"fmt"
	"github.com/puras/mog/config"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/inject"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/middleware"
	"github.com/puras/mog/web"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func Run(ctx context.Context, handler func(ctx context.Context) (func(), error)) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	cleanFunc, err := handler(ctx)
	if err != nil {
		return err
	}

EXIT:
	for {
		sig := <-sc
		logger.Context(ctx).Info("Receive signal", zap.String("signal", sig.String()))
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			state = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	logger.Context(ctx).Info("Server exit, bye...")
	time.Sleep(time.Millisecond * 100)
	os.Exit(state)
	return nil
}

func Start(ctx context.Context, injector *inject.Injector, registryRoutes func(ctx context.Context, e *gin.Engine) error, parseCurrentUser func(c *gin.Context) (string, error)) (func(), error) {
	logger.Context(ctx).Info("Start...")

	clean, err := startHTTPServer(ctx, registryRoutes, parseCurrentUser)
	if err != nil {
		return nil, err
	}

	return func() {
		if clean != nil {
			clean()
		}
	}, nil
}

func startHTTPServer(ctx context.Context, registryRoutes func(ctx context.Context, e *gin.Engine) error, parseCurrentUser func(c *gin.Context) (string, error)) (func(), error) {
	gin.SetMode(gin.DebugMode)

	e := gin.New()

	// 中间件应用
	e.Use(middleware.Recovery())
	e.Use(middleware.Trace())
	e.Use(middleware.Logger())
	e.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: []string{"/api/v1"},
		SkippedPathPrefixes: config.C.Middleware.Auth.SkippedPathPrefixes,
		ParseUser:           parseCurrentUser,
	}))

	e.GET("/health", func(c *gin.Context) {
		web.ResOk(c)
	})

	e.NoMethod(func(c *gin.Context) {
		web.ResError(c, errors.MethodNotAllowed("", "Method not allowed"))
	})

	e.NoRoute(func(c *gin.Context) {
		web.ResError(c, errors.NotFound("", "Not found"))
	})

	if registryRoutes != nil {
		err := registryRoutes(ctx, e)
		if err != nil {
			return nil, err
		}
	}

	serv := &http.Server{
		Addr:         config.C.General.HTTP.Addr,
		Handler:      e,
		ReadTimeout:  time.Second * time.Duration(config.C.General.HTTP.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(config.C.General.HTTP.WriteTimeout),
		IdleTimeout:  time.Second * time.Duration(config.C.General.HTTP.IdleTimeout),
	}
	logger.Context(ctx).Info(fmt.Sprintf("HTTP server is listening on %s", serv.Addr))
	go func() {
		err := serv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Context(ctx).Error("Failed to listen http server", zap.Error(err))
			panic(err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.C.General.HTTP.ShutdownTimeout))
		defer cancel()

		serv.SetKeepAlivesEnabled(false)
		if err := serv.Shutdown(ctx); err != nil {
			logger.Context(ctx).Error("Failed to shutdown http server", zap.Error(err))
		}
	}, nil
}
