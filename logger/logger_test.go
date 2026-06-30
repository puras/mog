package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// captureWriter 把写入的字节流转到 buffer，便于断言。
type captureWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *captureWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}
func (w *captureWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}
func (w *captureWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Reset()
}

// installJSONLogger 替换全局 zap logger，测试结束后还原。
func installJSONLogger(w *captureWriter, level zapcore.Level) func() {
	enc := zapcore.NewJSONEncoder(jsonEncoderConfig(0))
	core := zapcore.NewCore(enc, zapcore.AddSync(w), level)
	l := zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
	SetGlobal(l)
	return func() { SetGlobal(nil) }
}

func TestFrom_AutoInjectCtxFields(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	ctx := NewTraceId(context.Background(), "trace-xyz")
	ctx = NewUserId(ctx, "user-42")
	ctx = NewTag(ctx, TagKeyRequest)
	From(ctx).Info("hello", "k", "v")

	line := strings.TrimSpace(w.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("invalid json: %s\n", line)
	}
	if m["trace_id"] != "trace-xyz" || m["user_id"] != "user-42" || m["tag"] != TagKeyRequest {
		t.Fatalf("missing ctx fields: %+v", m)
	}
	if m["msg"] != "hello" || m["k"] != "v" {
		t.Fatalf("payload mismatch: %+v", m)
	}
}

func TestSpan_NestFieldsAndParent(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	rootCtx := context.Background()
	rootCtx, root := Start(rootCtx, "root")
	defer root.End()

	childCtx, child := Start(rootCtx, "child")
	_ = childCtx
	child.Set("user_id", "alice")
	child.End()

	root.End()

	lines := strings.Split(strings.TrimSpace(w.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("want >=2 lines, got %d: %s", len(lines), w.String())
	}
	// child should carry span_parent=root.
	var childMap map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &childMap); err != nil {
		t.Fatalf("child json: %s", lines[0])
	}
	if childMap["span"] != "child" || childMap["span_parent"] != "root" {
		t.Fatalf("child span/parent wrong: %+v", childMap)
	}
	if childMap["user_id"] != "alice" {
		t.Fatalf("child user_id missing: %+v", childMap)
	}
	if childMap["span_depth"] != float64(1) {
		t.Fatalf("child depth want 1, got %v", childMap["span_depth"])
	}
}

func TestSpan_ErrorPromotesLevel(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	ctx, sp := Start(context.Background(), "boom")
	_ = ctx
	sp.Err(errBoom{}).End()

	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(w.String())), &m); err != nil {
		t.Fatal(err)
	}
	if m["level"] != "error" {
		t.Fatalf("expected error level, got %v", m["level"])
	}
	if m["error"] == nil {
		t.Fatalf("expected error field, got %+v", m)
	}
}

func TestMultiCore_DisableSink(t *testing.T) {
	w1, w2 := &captureWriter{}, &captureWriter{}
	enc := zapcore.NewJSONEncoder(jsonEncoderConfig(0))
	c1 := zapcore.NewCore(enc, zapcore.AddSync(w1), zapcore.DebugLevel)
	c2 := zapcore.NewCore(enc, zapcore.AddSync(w2), zapcore.DebugLevel)

	tee := Multi(c1, c2)
	l := zap.New(tee, zap.AddCallerSkip(1))
	SetGlobal(l)
	defer SetGlobal(nil)

	From(context.Background()).Info("a")
	if !strings.Contains(w1.String(), "a") || !strings.Contains(w2.String(), "a") {
		t.Fatalf("want both cores filled, w1=%q w2=%q", w1.String(), w2.String())
	}

	if !DisableSink(tee, 0) {
		t.Fatal("expected sink disabled")
	}
	w1.Reset()
	w2.Reset()
	From(context.Background()).Info("b")
	if strings.Contains(w1.String(), "b") {
		t.Fatal("disabled core should not receive b")
	}
	if !strings.Contains(w2.String(), "b") {
		t.Fatal("active core should receive b")
	}
}
