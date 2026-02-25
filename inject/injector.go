package inject

import (
	"context"

	"github.com/puras/mog/dbx"
	"github.com/puras/mog/jwtx"
	"gorm.io/gorm"
)

// Injector 依赖注入器
type Injector struct {
	DB    *gorm.DB
	Auth  jwtx.Auth
	Trans *dbx.Trans
}

// InitInjector 初始化注入器（手动依赖注入）
// 返回 Injector 实例和清理函数
func InitInjector(ctx context.Context) (*Injector, func(), error) {
	var cleanFns []func()

	// 1. 初始化数据库
	db, dbClean, err := dbx.InitDB(ctx)
	if err != nil {
		return nil, nil, err
	}
	if dbClean != nil {
		cleanFns = append(cleanFns, dbClean)
	}

	// 2. 初始化JWT认证
	auth, authClean, err := jwtx.InitAuth(ctx)
	if err != nil {
		// 清理已初始化的资源
		for _, fn := range cleanFns {
			fn()
		}
		return nil, nil, err
	}
	if authClean != nil {
		cleanFns = append(cleanFns, authClean)
	}

	// 3. 构造注入器
	trans := &dbx.Trans{DB: db}
	injector := &Injector{
		DB:    db,
		Auth:  auth,
		Trans: trans,
	}

	// 4. 返回清理函数（逆序执行）
	return injector, func() {
		for i := len(cleanFns) - 1; i >= 0; i-- {
			cleanFns[i]()
		}
	}, nil
}
