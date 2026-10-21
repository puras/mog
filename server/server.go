package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/puras/mog/config"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/inject"
	"github.com/puras/mog/logger"
	"github.com/puras/mog/middleware"
	"github.com/puras/mog/web"

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
		logger.From(ctx).Info("Receive signal", zap.String("signal", sig.String()))
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
	logger.From(ctx).Info("Server exit, bye...")
	time.Sleep(time.Millisecond * 100)
	os.Exit(state)
	return nil
}

func Start(ctx context.Context, injector *inject.Injector, registryRoutes func(ctx context.Context, e *gin.Engine) error, parseCurrentUser func(c *gin.Context) (*middleware.AuthInfo, error)) (func(), error) {
	logger.From(ctx).Info("Start...")

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

// applyGinMode 让 gin 模式跟随 General.DebugMode：此前写死 DebugMode，
// 生产环境也会打印路由表与 [GIN-debug] 日志。与 command/start.go 的
// config.C.IsDebug() 判定保持同一开关。
func applyGinMode() {
	if config.C.IsDebug() {
		gin.SetMode(gin.DebugMode)
		return
	}
	gin.SetMode(gin.ReleaseMode)
}

func startHTTPServer(ctx context.Context, registryRoutes func(ctx context.Context, e *gin.Engine) error, parseCurrentUser func(c *gin.Context) (*middleware.AuthInfo, error)) (func(), error) {
	applyGinMode()

	e := gin.New()

	// 中间件应用：全部走 config，空值由各中间件的归一逻辑兜底。
	// 此前这里用的是 middleware.Recovery/Trace/Logger()，即写死的
	// Default*Config，导致 config 里 [Middleware.*] 整段配置无人消费。
	e.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		Skip: config.C.Middleware.Recovery.Skip,
	}))
	e.Use(middleware.TraceWithConfig(middleware.TraceConfig{
		SkippedPathPrefixes: config.C.Middleware.Trace.SkippedPathPrefixes,
		RequestHeaderKey:    config.C.Middleware.Trace.RequestHeaderKey,
		ResponseTraceKey:    config.C.Middleware.Trace.ResponseTraceKey,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		SkippedPathPrefixes:      config.C.Middleware.Logger.SkippedPathPrefixes,
		MaxOutputRequestBodyLen:  config.C.Middleware.Logger.MaxOutputRequestBodyLen,
		MaxOutputResponseBodyLen: config.C.Middleware.Logger.MaxOutputResponseBodyLen,
	}))
	e.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: []string{config.C.General.ContextPath},
		SkippedPathPrefixes: config.C.Middleware.Auth.SkippedPathPrefixes,
		Parse:               parseCurrentUser,
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
	logger.From(ctx).Info(fmt.Sprintf("HTTP server is listening on %s", serv.Addr))
	go func() {
		err := serv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.From(ctx).Error("Failed to listen http server", zap.Error(err))
			panic(err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.C.General.HTTP.ShutdownTimeout))
		defer cancel()

		serv.SetKeepAlivesEnabled(false)
		if err := serv.Shutdown(ctx); err != nil {
			logger.From(ctx).Error("Failed to shutdown http server", zap.Error(err))
		}
	}, nil
}
