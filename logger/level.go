// Package logger 提供 mog 框架的高性能日志能力。
//
// 设计要点：
//   - 一份代码同时输出到 console（彩色、单行紧凑、span 缩进）和 file（JSON、滚动日志）；
//   - 上下文自动串联 trace_id / user_id / tag / span；
//   - 通过 Start / SpanHandle 提供廉价的跨函数追踪能力；
//   - 业务侧调用保持 logger.From(ctx).Info(...) 风格不变。
package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

// ParseLevel 把字符串解析为 zapcore.Level，未识别时返回 InfoLevel 与错误。
func ParseLevel(s string) (zapcore.Level, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(s)))); err != nil {
		return zapcore.InfoLevel, fmt.Errorf("logger: unknown level %q: %w", s, err)
	}
	return lvl, nil
}
