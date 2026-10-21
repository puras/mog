package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/puras/mog/config"
)

// gin 模式必须跟随 General.DebugMode。此前写死 DebugMode，生产环境也会
// 打印路由表与 [GIN-debug] 日志。
func TestGinModeFollowsDebugMode(t *testing.T) {
	old := config.C.General.DebugMode
	defer func() {
		config.C.General.DebugMode = old
		gin.SetMode(gin.TestMode)
	}()

	tests := []struct {
		name  string
		debug bool
		want  string
	}{
		{"调试模式", true, gin.DebugMode},
		{"发布模式", false, gin.ReleaseMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.C.General.DebugMode = tt.debug
			applyGinMode()
			if got := gin.Mode(); got != tt.want {
				t.Errorf("gin.Mode() = %q, want %q", got, tt.want)
			}
		})
	}
}
