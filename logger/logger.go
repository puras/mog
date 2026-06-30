package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// callerFieldKey 是 Logger.Info/Warn/Error 自动注入的 field key，
// 业务层无需关心，console encoder / JSON encoder 都会识别。
const callerFieldKey = "_log_caller"

// callerFieldValue 携带现场 caller 信息（file:line）。
// 用 zapcore.ObjectMarshaler 让序列化可控。
type callerFieldValue struct {
	file string
	line int
}

func (c callerFieldValue) String() string {
	if c.file == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.file, c.line)
}

// MarshalLogObject 让 JSON encoder 直接拿到 file/line 两个字段。
func (c callerFieldValue) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file", c.file)
	enc.AddInt("line", c.line)
	return nil
}

// addCallerField 用 runtime.Caller 抓 caller，封装成 zap.Field。
// 从给定的 baseSkip 出发向上探测，找到第一个不在 logger 包内的 caller 帧。
// 这样无论 Logger.Info 上层包了几层都能准确定位业务调用方。
func addCallerField(baseSkip int) zapcore.Field {
	for skip := baseSkip; skip < baseSkip+10; skip++ {
		_, file, _, ok := runtime.Caller(skip)
		if !ok {
			return zap.Skip()
		}
		// 跳过自身：caller file 在 logger 包内就不是业务。
		if !strings.Contains(file, "/logger/") && !strings.HasSuffix(file, "/logger") {
			_, file, line, _ := runtime.Caller(skip)
			return zap.Object(callerFieldKey, callerFieldValue{file: file, line: line})
		}
	}
	return zap.Skip()
}

// extractCaller 从 fields 中拿出 callerField（如果有）。
func extractCaller(fields []zapcore.Field) (zapcore.Field, bool) {
	for _, f := range fields {
		if f.Key == callerFieldKey {
			return f, true
		}
	}
	return zapcore.Field{}, false
}

// 内部全局 zap logger 句柄，由 Init() 一次性安装。
var globalLogger atomic.Pointer[zap.Logger]

// SetGlobal 仅在测试或 fallback 场景使用。业务应调用 Init。
func SetGlobal(l *zap.Logger) {
	if l == nil {
		return
	}
	globalLogger.Store(l)
	zap.ReplaceGlobals(l)
}

// L 返回底层 zap logger，可能为 nil。
func L() *zap.Logger {
	return globalLogger.Load()
}

// From 返回带 ctx 字段的 *Logger，自动注入 trace_id / user_id / tag / span。
//
// 用法：
//
//	log := logger.From(ctx)
//	log.Info("dispatch ok", "channel", ch.ID, "ms", 12.3)
//	log.Error("upstream failed", logger.Err(err), "url", url)
func From(ctx context.Context) *Logger {
	base := L()
	if base == nil {
		base = zap.NewNop()
	}
	return &Logger{ctx: ctx, z: loggerFrom(ctx)}
}

// Skip 调用 business 包时需要主动禁用 caller skip 时使用。
func (l *Logger) Skip(n int) *Logger {
	if l == nil {
		return nil
	}
	l.z = l.z.WithOptions(zap.AddCallerSkip(n))
	return l
}

// Logger 是 zap.Logger 的轻包装，提供"key, value"风格的便利 API（slog 风格）。
type Logger struct {
	ctx context.Context
	z   *zap.Logger
}

// S 兼容旧 mog 的 logger.Context(ctx) 调用。
func S(ctx context.Context) *Logger { return From(ctx) }

// Context 兼容老 mog 调用点：logger.Context(ctx).Info(...) 等价于 logger.From(ctx).Info(...)。
func Context(ctx context.Context) *Logger { return From(ctx) }

// Info 打印 info 级日志。同时支持两种调用风格：
//   - zap 风格：Info(msg, zap.String(...), zap.Int(...))，全部 args 都是 zap.Field；
//   - slog 风格：Info(msg, "key", value, "key2", value2)，args 按 key/value 解析。
//
// 自动注入 caller field，encoder 读这个 field 输出 file:line。
// 不依赖 zap 的 WithCaller，因此 logger.CallerSkip 不需要业务配置。
func (l *Logger) Info(msg string, args ...any) {
	if l == nil || l.z == nil {
		return
	}
	caller := addCallerField(2) // [addCallerField -> Info -> 业务]
	if fields, ok := tryAllFields(args); ok {
		fields = append([]zapcore.Field{caller}, fields...)
		l.z.Info(msg, fields...)
		return
	}
	extra := argsToFields(args)
	extra = append([]zapcore.Field{caller}, extra...)
	l.z.Info(msg, extra...)
}

// Warn 同 Info。
func (l *Logger) Warn(msg string, args ...any) {
	if l == nil || l.z == nil {
		return
	}
	caller := addCallerField(2)
	if fields, ok := tryAllFields(args); ok {
		fields = append([]zapcore.Field{caller}, fields...)
		l.z.Warn(msg, fields...)
		return
	}
	extra := argsToFields(args)
	extra = append([]zapcore.Field{caller}, extra...)
	l.z.Warn(msg, extra...)
}

// Error 同 Info。
func (l *Logger) Error(msg string, args ...any) {
	if l == nil || l.z == nil {
		return
	}
	caller := addCallerField(2)
	if fields, ok := tryAllFields(args); ok {
		fields = append([]zapcore.Field{caller}, fields...)
		l.z.Error(msg, fields...)
		return
	}
	extra := argsToFields(args)
	extra = append([]zapcore.Field{caller}, extra...)
	l.z.Error(msg, extra...)
}

// Fatal 同 Info。
func (l *Logger) Fatal(msg string, args ...any) {
	if l == nil || l.z == nil {
		return
	}
	caller := addCallerField(2)
	if fields, ok := tryAllFields(args); ok {
		fields = append([]zapcore.Field{caller}, fields...)
		l.z.Fatal(msg, fields...)
		return
	}
	extra := argsToFields(args)
	extra = append([]zapcore.Field{caller}, extra...)
	l.z.Fatal(msg, extra...)
}

// Debug 同 Info。
func (l *Logger) Debug(msg string, args ...any) {
	if l == nil || l.z == nil {
		return
	}
	caller := addCallerField(2)
	if fields, ok := tryAllFields(args); ok {
		fields = append([]zapcore.Field{caller}, fields...)
		l.z.Debug(msg, fields...)
		return
	}
	extra := argsToFields(args)
	extra = append([]zapcore.Field{caller}, extra...)
	l.z.Debug(msg, extra...)
}

// tryAllFields 检查 args 是否全部为 zap.Field；是则直接转 []zap.Field 透传。
// 这是为了兼容老 mog 中 "Logger.Info(msg, fields...)" 的调用风格，
// 避免被误识别为 slog key/value 解析导致 Field 结构体被反射输出。
func tryAllFields(args []any) ([]zap.Field, bool) {
	if len(args) == 0 {
		return nil, true
	}
	out := make([]zap.Field, 0, len(args))
	for _, a := range args {
		f, ok := a.(zapcore.Field)
		if !ok {
			return nil, false
		}
		out = append(out, f)
	}
	return out, true
}

// With 派生带附加字段的 Logger（不影响源 logger）。
func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{ctx: l.ctx, z: l.z.With(argsToFields(args)...)}
}

// Err 辅助：把任意 error 转为 zap.Error 字段。
// 与直接传 error 不同，Err 允许 nil 不入参。
func Err(err error) zap.Field {
	return zap.Error(err)
}

// InfoF 等同 Info(msg, fields...) —— 给持有 []zap.Field 的调用点使用。
// 老代码最常见的写法就是 logger.Context(ctx).Info(msg, fields...)，
// 保留这组方法可避免大量手工展开。
func (l *Logger) InfoF(msg string, fields ...zap.Field) {
	if l == nil || l.z == nil {
		return
	}
	l.z.Info(msg, fields...)
}

func (l *Logger) DebugF(msg string, fields ...zap.Field) {
	if l == nil || l.z == nil {
		return
	}
	l.z.Debug(msg, fields...)
}

func (l *Logger) WarnF(msg string, fields ...zap.Field) {
	if l == nil || l.z == nil {
		return
	}
	l.z.Warn(msg, fields...)
}

func (l *Logger) ErrorF(msg string, fields ...zap.Field) {
	if l == nil || l.z == nil {
		return
	}
	l.z.Error(msg, fields...)
}

// argsToFields 将 key/value 风格的 args 转 zap.Field 列表。
// 奇数个元素把最后一个当作 value 并以 key="" 形式追加；
//   - 当 value 是 zap.Field 时直接透传；
//   - 当 value 是 error 时转 zap.Error（不会因 nil panic）；
//   - 其余走 anyToField。
func argsToFields(args []any) []zap.Field {
	if len(args) == 0 {
		return nil
	}
	out := make([]zap.Field, 0, len(args)/2+1)
	for i := 0; i < len(args); i += 2 {
		var (
			k string
			v any
		)
		if i+1 < len(args) {
			k, _ = args[i].(string)
			v = args[i+1]
		} else {
			k = fmt.Sprintf("%v", args[i])
		}
		if k == "" {
			continue
		}
		if f, ok := v.(zap.Field); ok {
			out = append(out, f)
			continue
		}
		// error 特化：让 From(ctx).Info(..., "err", err) 更顺手
		if e, ok := v.(error); ok {
			if e == nil {
				continue
			}
			out = append(out, zap.Error(e))
			continue
		}
		out = append(out, anyToField(k, v))
	}
	return out
}
