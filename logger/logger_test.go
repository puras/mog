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
	// caller 字段由 Logger.Info 自动注入，结构是 {file, line}。
	caller, ok := m[callerFieldKey].(map[string]any)
	if !ok {
		t.Fatalf("caller field missing or wrong type: %+v", m[callerFieldKey])
	}
	// 测试场景下 testing 框架会插一帧，所以 file 不一定是 logger_test.go。
	// 只要非空、非 logger 包内，且 line > 0 即可。
	if caller["file"].(string) == "" {
		t.Fatalf("caller.file empty")
	}
	if strings.Contains(caller["file"].(string), "/logger/") {
		t.Fatalf("caller.file should be outside logger pkg, got %v", caller["file"])
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

// 父 span End 之后，已启动的子 span 仍须打印正确的 span_parent。
//
// 修复前 span 走 sync.Pool 复用：父 End 时被归还，下一次 Start 取出同一
// 对象并写上新名字，子 span 的 span_parent 就会变成别人的名字。
func TestSpan_ParentNameSurvivesParentEnd(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	ctx := context.Background()
	ctxRoot, root := Start(ctx, "root")
	_, child := Start(ctxRoot, "child")

	root.End() // 修复前此处会把 root 归还池中

	// 另起一条无关调用链：若走对象池，它会拿走刚归还的 root 对象
	_, stranger := Start(ctx, "STRANGER")

	child.End()
	stranger.End()

	var childEntry map[string]any
	for _, line := range strings.Split(strings.TrimSpace(w.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("解析日志失败: %v (%s)", err, line)
		}
		if m["span"] == "child" {
			childEntry = m
		}
	}
	if childEntry == nil {
		t.Fatalf("未找到 child span 日志: %s", w.String())
	}
	if got := childEntry["span_parent"]; got != "root" {
		t.Errorf("child 的 span_parent = %v，want root（父 span 已被复用）", got)
	}
}

// 同一 goroutine 内 span 字段互不串味：修复前池复用会让 fields 残留。
func TestSpan_FieldsDoNotLeak(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	ctx := context.Background()
	_, first := Start(ctx, "first")
	first.Set("leak", "should-not-appear")
	first.End()

	_, second := Start(ctx, "second")
	second.End()

	for _, line := range strings.Split(strings.TrimSpace(w.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("解析日志失败: %v", err)
		}
		if m["span"] == "second" {
			if _, ok := m["leak"]; ok {
				t.Errorf("second span 残留了 first 的字段: %+v", m)
			}
		}
	}
}

// 并发开 span 不应触发数据竞争（配合 -race 使用）。
func TestSpan_Concurrent(t *testing.T) {
	w := &captureWriter{}
	defer installJSONLogger(w, zapcore.DebugLevel)()

	ctx := context.Background()
	ctxRoot, root := Start(ctx, "root")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, sp := Start(ctxRoot, "child")
			sp.Set("k", "v")
			sp.End()
		}()
	}
	wg.Wait()
	root.End()

	if n := root.s.counter.Load(); n != 8 {
		t.Errorf("span_children = %d, want 8", n)
	}
}
