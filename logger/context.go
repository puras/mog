package logger

import (
	"context"

	"go.uber.org/zap"
)

// 公共 ctx key —— 全部导出，便于业务包直接 context.WithValue 复用。
type (
	// TraceIdKey 存放链路 trace id。
	TraceIdKey struct{}
	// UserIdKey 存放当前用户 id。
	UserIdKey struct{}
	// TagKey 存放调用域分类，如 request / recovery / system。
	TagKey struct{}
	// SpanKey 存放当前 goroutine 链上的根 span。
	SpanKey struct{}
)

// 预置 tag 取值 —— 业务可直接 context.WithValue(... TagKey, logger.TagKeyRequest)。
const (
	TagKeyRequest  = "request"
	TagKeyRecovery = "recovery"
	TagKeySystem   = "system"
	TagKeyLogin    = "login"
	TagKeyLogout   = "logout"
	TagKeyOperate  = "operate"
	TagKeyMain     = "main"
)

// NewTraceId 把 trace id 注入 ctx。
func NewTraceId(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, TraceIdKey{}, id)
}

// FromTraceId 读取 trace id，未设置时返回空串。
func FromTraceId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(TraceIdKey{}).(string); ok {
		return v
	}
	return ""
}

// NewUserId 把 user id 注入 ctx。
func NewUserId(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, UserIdKey{}, id)
}

// FromUserId 读取 user id，未设置时返回空串。
func FromUserId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(UserIdKey{}).(string); ok {
		return v
	}
	return ""
}

// NewTag 把 tag 注入 ctx。常量见 TagKeyRequest 等。
func NewTag(ctx context.Context, tag string) context.Context {
	if tag == "" {
		return ctx
	}
	return context.WithValue(ctx, TagKey{}, tag)
}

// FromTag 读取 tag。
func FromTag(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(TagKey{}).(string); ok {
		return v
	}
	return ""
}

// newSpanCtx 内部 helper：把 span 写入 ctx。
func newSpanCtx(ctx context.Context, s *span) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, SpanKey{}, s)
}

// fromSpan 读取 ctx 上的 span，未设置返回 nil。
func fromSpan(ctx context.Context) *span {
	if ctx == nil {
		return nil
	}
	if v, ok := ctx.Value(SpanKey{}).(*span); ok {
		return v
	}
	return nil
}

// LoggerFrom 用于基础设施（middleware/handler）一次性把所有 ctx 字段包装为 zap.Logger。
// 不建议业务直接使用，业务请用 From(ctx)。
func loggerFrom(ctx context.Context) *zap.Logger {
	base := zap.L()
	if base == nil {
		return zap.NewNop().Sugar().Desugar()
	}
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
	if s := fromSpan(ctx); s != nil {
		fields = append(fields,
			zap.String("span", s.name),
			zap.Int32("span_depth", s.depth),
			zap.String("span_parent", parentName(s)),
		)
	}
	return base.With(fields...)
}

func parentName(s *span) string {
	if s.parent == nil {
		return ""
	}
	return s.parent.name
}
