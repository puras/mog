package logger

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// spanPool 通过 sync.Pool 复用 span，零分配创建新 span。
var spanPool = sync.Pool{
	New: func() any { return &span{} },
}

// span 是单次调用链上的逻辑节点，承载 name / parent / depth / start / fields / err。
type span struct {
	name    string
	parent  *span
	depth   int32
	start   time.Time
	fields  []zapcore.Field
	err     error
	ended   atomic.Bool
	counter atomic.Int32
}

// Start 开一段 span；返回新的 ctx 和 handle，defer handle.End() 即可。
//
// 用法：
//
//	ctx, sp := logger.Start(ctx, "channel.proxy")
//	defer sp.End()
//	if err != nil { sp.Set("err", err); return err }
func Start(ctx context.Context, name string) (context.Context, *SpanHandle) {
	s := spanPool.Get().(*span)
	s.name = name
	s.start = time.Now()
	s.err = nil
	s.ended.Store(false)
	s.counter.Store(0)
	s.fields = s.fields[:0]

	if p := fromSpan(ctx); p != nil {
		s.parent = p
		s.depth = p.depth + 1
		p.counter.Add(1)
	} else {
		s.parent = nil
		s.depth = 0
	}

	return newSpanCtx(ctx, s), &SpanHandle{s: s}
}

// SpanHandle 提供给业务来打点的句柄；不暴露 span 内部状态。
type SpanHandle struct {
	s *span
}

// End 结束 span 并打印日志。多次调用安全，仅第一次生效。
func (h *SpanHandle) End() {
	if h == nil || h.s == nil {
		return
	}
	if !h.s.ended.CompareAndSwap(false, true) {
		return
	}

	cost := time.Since(h.s.start)
	fields := make([]zapcore.Field, 0, len(h.s.fields)+5)
	fields = append(fields,
		zap.String("span", h.s.name),
		zap.Int32("span_depth", h.s.depth),
		zap.Int32("span_children", h.s.counter.Load()),
		zap.Int64("cost_ms", cost.Milliseconds()),
	)
	if h.s.parent != nil {
		fields = append(fields, zap.String("span_parent", h.s.parent.name))
	}
	if h.s.err != nil {
		fields = append(fields, zap.Error(h.s.err))
	}
	for _, f := range h.s.fields {
		fields = append(fields, f)
	}

	base := zap.L()
	if base == nil {
		spanPool.Put(h.s)
		return
	}
	if h.s.err != nil {
		base.Error(h.s.name, fields...)
	} else {
		base.Info(h.s.name, fields...)
	}
	spanPool.Put(h.s)
}

// Set 在 span 上附加 KV。k 必须为常量字符串；v 自动转 zap.Field。
func (h *SpanHandle) Set(k string, v any) *SpanHandle {
	if h == nil || h.s == nil {
		return h
	}
	h.s.fields = append(h.s.fields, anyToField(k, v))
	return h
}

// Err 标记 span 失败。重复调用不覆盖前值。
func (h *SpanHandle) Err(err error) *SpanHandle {
	if h == nil || h.s == nil {
		return h
	}
	if err != nil && h.s.err == nil {
		h.s.err = err
	}
	return h
}

// Errf 便捷字符串错误包装。
func (h *SpanHandle) Errf(format string, args ...any) *SpanHandle {
	return h.Err(errorStringf(format, args...))
}

// type errorString 延迟到 errors.go（避免 span.go import cycle）。
type errorString string

func (e errorString) Error() string { return string(e) }

func errorStringf(format string, args ...any) errorString {
	return errorString(format2(format, args...))
}

// anyToField 把任意值转 zap.Field；规则如下：
//
//	nil       → zap.Any(k, nil)
//	string    → zap.String(k, v)
//	[]byte    → zap.String(k, string(b))
//	int/int64 → zap.Int64(k, int64(v))
//	bool      → zap.Bool(k, v)
//	error     → zap.Error(v)
//	其余       → zap.Any(k, v)（zap 会反射，必要时再调专门的构造函数）
func anyToField(k string, v any) zapcore.Field {
	switch x := v.(type) {
	case nil:
		return zap.Any(k, nil)
	case string:
		return zap.String(k, x)
	case []byte:
		return zap.String(k, string(x))
	case bool:
		return zap.Bool(k, x)
	case error:
		return zap.Error(x)
	case int:
		return zap.Int64(k, int64(x))
	case int8:
		return zap.Int64(k, int64(x))
	case int16:
		return zap.Int64(k, int64(x))
	case int32:
		return zap.Int64(k, int64(x))
	case int64:
		return zap.Int64(k, x)
	case uint:
		return zap.Uint64(k, uint64(x))
	case uint32:
		return zap.Uint64(k, uint64(x))
	case uint64:
		return zap.Uint64(k, x)
	case float32:
		return zap.Float64(k, float64(x))
	case float64:
		return zap.Float64(k, x)
	case time.Time:
		return zap.Time(k, x)
	case time.Duration:
		return zap.Duration(k, x)
	default:
		return zap.Any(k, x)
	}
}

// format2 避免 errors.go 提前加载 fmt.Errorf 链路：纯 sprintf。
func format2(format string, args ...any) string {
	// 使用 fmt.Sprintf 以避免重复造轮子
	return sprintf(format, args...)
}
