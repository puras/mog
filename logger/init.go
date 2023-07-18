package logger

import (
	"context"
	"github.com/puras/mog/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(ctx context.Context) (func(), error) {
	var conf zap.Config
	if config.C.Logger.Debug {
		config.C.Logger.Level = "debug"
		conf = zap.NewDevelopmentConfig()
	} else {
		conf = zap.NewProductionConfig()
	}
	level, err := zapcore.ParseLevel(config.C.Logger.Level)
	if err != nil {
		return nil, err
	}
	conf.Level.SetLevel(level)

	var (
		logger   *zap.Logger
		cleanFns []func()
	)
	if config.C.Logger.File.Enable {
		filename := config.C.Logger.File.Path
		_ = os.MkdirAll(filepath.Dir(filename), 0777)
		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    config.C.Logger.File.MaxSize,
			MaxBackups: config.C.Logger.File.MaxBackups,
			Compress:   false,
			LocalTime:  true,
		}

		cleanFns = append(cleanFns, func() {
			_ = fileWriter.Close()
		})

		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(conf.EncoderConfig),
			zapcore.AddSync(fileWriter),
			conf.Level,
		)
		logger = zap.New(zc)
	} else {
		ilog, err := conf.Build()
		if err != nil {
			return nil, err
		}
		logger = ilog
	}
	skip := config.C.Logger.CallerSkip
	if skip <= 0 {
		skip = 2
	}
	logger = logger.WithOptions(
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(skip),
	)
	zap.ReplaceGlobals(logger)
	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}
