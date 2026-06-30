package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// mux 保护 Init 幂等。
var mux sync.Mutex

// Init 装载默认配置（info 级别、console on、file off），并替换 zap.Global。
// 多次调用安全；后调一次覆盖前一次的全局 logger 与文件句柄。
func Init() (func(), error) {
	return InitWithConfig(defaultConfig())
}

// InitWithConfig 用指定 Config 装载。返回 shutdown() 关闭文件句柄。
func InitWithConfig(cfg Config) (func(), error) {
	mux.Lock()
	defer mux.Unlock()

	if cfg.CallerSkip <= 0 {
		cfg.CallerSkip = 2
	}

	// —— Color 设置 ——
	setColorMode(cfg.Console.Color)

	// —— 构造 console core ——
	var (
		cores    []zapcore.Core
		cleanups []func()
	)
	if cfg.Console.Enable {
		consoleCore := buildConsoleCore(cfg)
		cores = append(cores, consoleCore)
	}

	// —— 构造 file core ——
	if cfg.File.Enable {
		fileCore, cleanup, err := buildFileCore(cfg)
		if err != nil {
			return nil, err
		}
		if fileCore != nil {
			cores = append(cores, fileCore)
			if cleanup != nil {
				cleanups = append(cleanups, cleanup)
			}
		}
	}

	if len(cores) == 0 {
		// 没有任何 sink：装一个 Nop 防止业务 panic。
		cores = append(cores, zapcore.NewNopCore())
	}

	tee := Multi(cores...)
	l := zap.New(
		tee,
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(cfg.CallerSkip),
	)
	if cfg.Sampling > 0 {
		l = l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(c, time.Second, cfg.Sampling, cfg.Sampling)
		}))
	}
	SetGlobal(l)

	return func() {
		for _, fn := range cleanups {
			fn()
		}
		// 释放内部资源。
		_ = tee.Sync()
	}, nil
}

func buildConsoleCore(cfg Config) zapcore.Core {
	lvl := cfg.Console.MinLevel
	if lvl == 0 {
		lvl = cfg.Level
	}
	enc := NewConsoleEncoder(cfg.Console.Theme, cfg.Console.TimeLayout)
	w := zapcore.Lock(zapcore.AddSync(os.Stdout))
	return zapcore.NewCore(enc, w, lvl)
}

func buildFileCore(cfg Config) (zapcore.Core, func(), error) {
	if cfg.File.Path == "" {
		return nil, nil, nil
	}
	if err := os.MkdirAll(dirOf(cfg.File.Path), 0o755); err != nil {
		return nil, nil, err
	}
	w := &lumberjack.Logger{
		Filename:   cfg.File.Path,
		MaxSize:    cfg.File.MaxSize,
		MaxBackups: cfg.File.MaxBackups,
		MaxAge:     cfg.File.MaxAge,
		Compress:   cfg.File.Compress,
		LocalTime:  cfg.File.LocalTime,
	}
	lvl := cfg.File.MinLevel
	if lvl == 0 {
		lvl = cfg.Level
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoderConfig(cfg.CallerSkip)),
		zapcore.AddSync(w),
		lvl,
	)
	return core, func() { _ = w.Close() }, nil
}

func dirOf(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}
