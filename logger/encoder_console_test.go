package logger

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestConsoleEncoder_NoColorOutput(t *testing.T) {
	setColorMode("off")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{
			Time:    time.Date(2026, 6, 30, 14, 23, 5, 0, time.UTC),
			Level:   zapcore.InfoLevel,
			Message: "hello",
		},
		[]zapcore.Field{
			zap.String("span", "root"),
			zap.Int32("span_depth", 0),
			zap.String("user_id", "u1"),
			zap.String("extra", "value"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	if strings.ContainsAny(out, "\x1b") {
		t.Fatalf("expected no ANSI when color off: %q", out)
	}
	if !strings.Contains(out, "INFO") || !strings.Contains(out, "root") || !strings.Contains(out, "extra=value") {
		t.Fatalf("missing required fragments: %q", out)
	}
}

func TestConsoleEncoder_MessageColored(t *testing.T) {
	setColorMode("on")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.InfoLevel, Message: "[downstream]"},
		[]zapcore.Field{},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	// 标签 [downstream] 应当用 cyan+bold；空 rest 不写出颜色。
	if !strings.Contains(out, ansiCyan+ansiBold+"[downstream]"+ansiReset) {
		t.Fatalf("expected [downstream] tagged with cyan+bold, got %q", out)
	}
}

func TestConsoleEncoder_MessagePrefixAndRest(t *testing.T) {
	setColorMode("on")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.InfoLevel, Message: "[downstream] 请求 (claude-code → MoGo)"},
		[]zapcore.Field{},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	// [downstream] 用 cyan+bold
	if !strings.Contains(out, ansiCyan+ansiBold+"[downstream]"+ansiReset) {
		t.Fatalf("expected [downstream] in cyan+bold, got %q", out)
	}
	// 正文部分不应被包任何 ANSI 颜色（保持默认终端色）。
	if !strings.Contains(out, "请求 (claude-code → MoGo)") {
		t.Fatalf("expected raw rest text, got %q", out)
	}
	// 断言：ANSI 转义只出现在标签段；正文区段裸出。
	// 查找 rest 段前最近的换行或开头，验证它没前缀 ANSI。
	if strings.Contains(out, "\x1b"+ansiGray+"请求") {
		t.Fatalf("rest should not be wrapped in ansiGray, got %q", out)
	}
}

func TestConsoleEncoder_MessagePlainWhenColorOff(t *testing.T) {
	setColorMode("off")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.InfoLevel, Message: "[downstream] rest"},
		[]zapcore.Field{},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	if strings.ContainsAny(out, "\x1b") {
		t.Fatalf("no ANSI expected, got %q", out)
	}
	if !strings.Contains(out, "[downstream] rest") {
		t.Fatalf("raw message must be preserved, got %q", out)
	}
}

func TestConsoleEncoder_ColorOnForLevel(t *testing.T) {
	setColorMode("on")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.ErrorLevel, Message: "x"},
		[]zapcore.Field{},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	if !strings.Contains(out, "\x1b[31m") {
		t.Fatalf("expected red ANSI for ERROR: %q", out)
	}
}

func TestConsoleEncoder_IndentForSpan(t *testing.T) {
	setColorMode("off")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.InfoLevel, Message: "ignored msg"},
		[]zapcore.Field{
			zap.String("span", "child"),
			zap.Int32("span_depth", 2),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	if !strings.Contains(out, "  └─ child") {
		t.Fatalf("expected indent for span_depth=2, got %q", out)
	}
	if strings.Contains(out, "ignored msg") {
		t.Fatalf("span name should override msg: %q", out)
	}
}

func TestConsoleEncoder_FrameKeysSuppressed(t *testing.T) {
	setColorMode("off")
	defer setColorMode("auto")

	enc := NewConsoleEncoder("default", "15:04:05.000")
	buf, err := enc.EncodeEntry(
		zapcore.Entry{Level: zapcore.InfoLevel, Message: "m"},
		[]zapcore.Field{
			zap.String("trace_id", "tx"),
			zap.String("span", "x"),
			zap.String("extra", "v"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Free()
	out := buf.String()
	if strings.Contains(out, "trace_id=") {
		t.Fatalf("trace_id should be hidden by console encoder, got %q", out)
	}
	if !strings.Contains(out, "extra=v") {
		t.Fatalf("extra=v should be present, got %q", out)
	}
}

func TestKVValueTypes(t *testing.T) {
	// 保护 anyToField，避免回归。
	if f := anyToField("k", errBoom{}); f.Type != zapcore.ErrorType {
		t.Fatalf("anyToField err: type=%v", f.Type)
	}
	if f := anyToField("k", "v"); f.Type != zapcore.StringType {
		t.Fatalf("anyToField string: type=%v", f.Type)
	}
	if f := anyToField("k", int64(7)); f.Type != zapcore.Int64Type {
		t.Fatalf("anyToField int64: type=%v", f.Type)
	}
	if f := anyToField("k", nil); f.Type == 0 {
		t.Fatalf("anyToField nil must emit a field")
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }
