package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// jsonEncoderConfig 返回文件侧 JSON 编码器配置。
// 通用：保留 caller、stack、ISO8601 时间，便于 ELK / Loki 直接索引。
func jsonEncoderConfig(skip int) zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		TimeKey:        "ts",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// newJSONCore 直接返回一个 JSON 编码的 zapcore.Core，方便测试复用。
func newJSONCore(w zapcore.WriteSyncer, lvl zapcore.Level) zapcore.Core {
	enc := zapcore.NewJSONEncoder(jsonEncoderConfig(2))
	return zapcore.NewCore(enc, w, lvl)
}

// newFileCore 构造 lumberjack 滚动日志 + JSON。
func newFileCore(sink FileSink, lvl zapcore.Level, callerSkip int) (zapcore.Core, func(), error) {
	if !sink.Enable || sink.Path == "" {
		return nil, nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(sink.Path), 0o755); err != nil {
		return nil, nil, err
	}

	w := &lumberjack.Logger{
		Filename:   sink.Path,
		MaxSize:    sink.MaxSize,
		MaxBackups: sink.MaxBackups,
		MaxAge:     sink.MaxAge,
		Compress:   sink.Compress,
		LocalTime:  sink.LocalTime,
	}
	coreLevel := sink.MinLevel
	if coreLevel == 0 {
		coreLevel = lvl
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoderConfig(callerSkip)),
		zapcore.AddSync(w),
		coreLevel,
	)
	cleanup := func() { _ = w.Close() }
	return core, cleanup, nil
}

// 静态断言：检查 zapcore.Core 接口使用 OK。
var _ zapcore.Core = (*nopCore)(nil)

type nopCore struct{}

func (nopCore) Enabled(zapcore.Level) bool        { return false }
func (nopCore) With([]zapcore.Field) zapcore.Core { return nopCore{} }
func (nopCore) Check(_ zapcore.Entry, _ *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return nil
}
func (nopCore) Write(zapcore.Entry, []zapcore.Field) error { return nil }
func (nopCore) Sync() error                                { return nil }
