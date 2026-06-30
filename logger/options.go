// Package logger —— 高级配置入口。
//
// 通过 Init / InitWithConfig 一次性安装 console + file 两路 sink；
// 业务代码使用 From(ctx) 取得带 trace_id / user_id / tag / span 字段的 logger。
package logger

import (
	"go.uber.org/zap/zapcore"
)

// ConsoleSink 控制台输出配置。
type ConsoleSink struct {
	// Enable 关闭后 Init 不输出到 stdout。
	Enable bool
	// Color 取值 auto / on / off。
	Color string
	// Theme 取值 default / minimal / bright。
	Theme string
	// TimeLayout 默认 "15:04:05.000"。
	TimeLayout string
	// MinLevel 小于该 level 的条目不在此处输出（继承全局 level 时保持 zap.Level(0)）。
	MinLevel zapcore.Level
}

// FileSink 文件输出配置。
type FileSink struct {
	// Enable 关闭后 Init 不写文件。
	Enable bool
	// Path 文件路径，必须设置。
	Path string
	// MaxSize 单文件大小上限，MB。
	MaxSize int
	// MaxBackups 保留历史文件数。
	MaxBackups int
	// MaxAge 历史文件最大保留天数。
	MaxAge int
	// Compress 是否 gzip 压缩历史文件。
	Compress bool
	// LocalTime 滚动文件名是否带本地时间（默认 UTC）。
	LocalTime bool
	// MinLevel 文件专用 level。
	MinLevel zapcore.Level
}

// Config Init 用的全量配置。
type Config struct {
	// Level 全局最低 level。
	Level zapcore.Level
	// CallerSkip 调栈跳过层数。
	CallerSkip int
	// Sampling 每秒同 msg 最多打印条数；0 关闭。
	Sampling int

	Console ConsoleSink
	File    FileSink
}

// defaultConfig 提供一个安全默认：console on、file off、level=info。
func defaultConfig() Config {
	return Config{
		Level:      zapcore.InfoLevel,
		CallerSkip: 2,
		Console: ConsoleSink{
			Enable:     true,
			Color:      "auto",
			Theme:      "default",
			TimeLayout: "15:04:05.000",
		},
	}
}
