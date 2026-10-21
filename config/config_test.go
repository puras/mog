package config

import (
	"os"
	"path/filepath"
	"testing"
)

// default 标签的正确性依赖 reflect.StructTag 语法：值必须用双引号包裹，
// 否则 StructTag.Get 直接返回空串且不报错——默认值静默失效，只有 go vet
// 的 structtag 检查会提示。
//
// 回归用例：涵盖原先漏引号的 CopyBody.MaxContentLen（曾静默为 0）。
// 注意 MustLoad 用 sync.Once 守卫，进程内只真正加载一次，因此本用例
// 不与其他加载用例并存（config 包内仅此一个加载断言）。
func TestDefaultTagsApplied(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config.toml")
	// 只写一个字段，其余全部依赖 default 标签兜底
	if err := os.WriteFile(f, []byte("[General]\nAppName = \"probe\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	MustLoad(f)

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"General.HTTP.Addr", C.General.HTTP.Addr, ":8000"},
		{"General.HTTP.ReadTimeout", C.General.HTTP.ReadTimeout, 60},
		{"General.HTTP.ShutdownTimeout", C.General.HTTP.ShutdownTimeout, 10},
		{"Storage.DataBase.Enable", C.Storage.DataBase.Enable, true},
		{"Storage.DataBase.Type", C.Storage.DataBase.Type, "sqlite3"},
		{"Storage.DataBase.MaxOpenConns", C.Storage.DataBase.MaxOpenConns, 100},
		{"Storage.DataBase.MaxIdleConns", C.Storage.DataBase.MaxIdleConns, 50},
		{"Storage.Cache.Type", C.Storage.Cache.Type, "memory"},
		{"Middleware.Auth.SigningMethod", C.Middleware.Auth.SigningMethod, "HS512"},
		{"Middleware.Auth.Expired", C.Middleware.Auth.Expired, 86400},
		{"Middleware.Logger.MaxOutputRequestBodyLen", C.Middleware.Logger.MaxOutputRequestBodyLen, 1048576},
		{"Middleware.Logger.MaxOutputResponseBodyLen", C.Middleware.Logger.MaxOutputResponseBodyLen, 1048576},
		// 这一项原来是 `default:33554432`（漏引号），默认值静默失效为 0
		{"Middleware.CopyBody.MaxContentLen", C.Middleware.CopyBody.MaxContentLen, int64(33554432)},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}
