package logger

import "fmt"

// sprintf 作为 fmt.Sprintf 的薄封装，避免 span.go 间接循环导入 fmt。
func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
