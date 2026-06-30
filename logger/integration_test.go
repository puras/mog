package logger

import (
	"bytes"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 测试一把"真实业务"行为：zap.Bool + zap.String + zap.Int，输出用 console encoder
func TestIntegration_BoolAndIntKV(t *testing.T) {
	setColorMode("off")
	defer setColorMode("auto")

	w := &bytes.Buffer{}
	enc := NewConsoleEncoder("default", "15:04:05.000")
	core := zapcore.NewCore(enc, zapcore.AddSync(w), zapcore.InfoLevel)
	l := zap.New(core, zap.AddCallerSkip(1))

	l.Info("[dispatch] Provider 执行",
		zap.Bool("isStream", true),
		zap.String("handler", "anthropic"),
		zap.Int64("cost_ms", 32223),
	)

	t.Logf("output: %s", w.String())
}
