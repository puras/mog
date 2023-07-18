package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

const (
	TagKeyMain     = "main"
	TagKeyRecovery = "recovery"
	TagKeyRequest  = "request"
	TagKeyLogin    = "login"
	TagKeyLogout   = "logout"
	TagKeySystem   = "system"
	TagKeyOperate  = "operate"
)

type (
	ctxLoggerKey  struct{}
	ctxTraceIdKey struct{}
	ctxUserIdKey  struct{}
	ctxTagKey     struct{}
	ctxStackKey   struct{}
)

func InitWithConfig(ctx context.Context) (func(), error) {
	return nil, nil
}

func NewLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

func FromLogger(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxLoggerKey{})
	if v != nil {
		if vv, ok := v.(*zap.Logger); ok {
			return vv
		}
	}
	return zap.L()
}

func NewTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, ctxTraceIdKey{}, traceId)
}

func FromTraceId(ctx context.Context) string {
	v := ctx.Value(ctxTraceIdKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NewUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, ctxUserIdKey{}, userId)
}

func FromUserId(ctx context.Context) string {
	v := ctx.Value(ctxUserIdKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NewTag(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, ctxTagKey{}, tag)
}

func FromTag(ctx context.Context) string {
	v := ctx.Value(ctxTagKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func NewStack(ctx context.Context, stack error) context.Context {
	return context.WithValue(ctx, ctxStackKey{}, stack)
}

func FromStack(ctx context.Context) error {
	v := ctx.Value(ctxStackKey{})
	if v != nil {
		if s, ok := v.(error); ok {
			return s
		}
	}
	return nil
}

func Context(ctx context.Context) *zap.Logger {
	var fields []zap.Field
	if v := FromTraceId(ctx); v != "" {
		fields = append(fields, zap.String("trace_id", v))
	}
	if v := FromUserId(ctx); v != "" {
		fields = append(fields, zap.String("user_id", v))
	}
	if v := FromTag(ctx); v != "" {
		fields = append(fields, zap.String("tag", v))
	}
	if v := FromStack(ctx); v != nil {
		fields = append(fields, zap.Error(v))
	}
	return FromLogger(ctx).With(fields...)
}

type PrintLogger struct{}

func (l *PrintLogger) Printf(format string, args ...any) {
	zap.L().Info(fmt.Sprintf(format, args...))
}
