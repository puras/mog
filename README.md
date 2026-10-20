# MoG

Mooko Go Web 基础框架。基于 Gin、Gorm 等开源项目构建，用于快速开发 Web 应用。

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 特性

- **Web 框架**: 基于 [Gin](https://github.com/gin-gonic/gin) 的高性能 HTTP 路由
- **ORM**: 使用 [Gorm](https://github.com/go-gorm/gorm) 进行数据库操作
- **数据库支持**: MySQL、PostgreSQL、SQLite3
- **认证**: 基于 JWT 的认证系统，支持 Token 撤销
- **缓存**: 支持 Redis、Badger、内存缓存
- **对象存储**: MinIO 客户端集成
- **邮件**: 邮件发送功能
- **日志**: 基于 zap 的结构化日志
- **配置**: TOML 格式配置文件
- **中间件**: 恢复、追踪、日志、认证等常用中间件
- **错误处理**: 统一的错误处理和响应格式
- **CRUD**: 开箱即用的 CRUD 功能

## 快速开始

### 安装

```bash
go get github.com/puras/mog
```

### 构建和运行

```bash
# 克隆仓库
git clone https://github.com/puras/mog.git
cd mog

# 构建项目
go build

# 运行服务器（使用默认配置文件 conf/config.tom.toml）
./app start

# 指定配置文件运行
./app start --conf conf/config.toml

# 以守护进程模式运行
./app start --daemon

# 停止服务器
./app stop
```

## 项目结构

```
mog/
├── cachex/        # 缓存封装（Redis、Badger、Memory）
├── command/       # CLI 命令
├── config/        # 配置管理
├── contextx/      # 上下文封装
├── crud/          # CRUD 基础实现
├── crypto/        # 加密工具
├── dbx/           # 数据库扩展（GORM 封装、事务、分页）
├── errors/        # 错误处理
├── inject/        # 依赖注入
├── jwtx/          # JWT 认证
├── logger/        # 日志系统
├── mail/          # 邮件发送
├── middleware/    # Gin 中间件
├── model/         # 数据模型基类
├── module/        # 模块系统
├── oss/           # 对象存储（MinIO）
├── server/        # 服务器启动
├── utils/         # 工具函数
└── web/           # Web 响应封装
```

## 核心模块

### 配置管理

使用 TOML 格式配置文件：

```toml
[General]
AppName = "myapp"
DebugMode = true

[General.HTTP]
Addr = ":8000"

[Storage.DataBase]
Type = "sqlite3"
DSN = "data/sqlite/app.db"
AutoMigrate = true

[Middleware.Auth]
SkippedPathPrefixes = ["/health", "/api/v1/login"]
```

### 依赖注入

手动依赖注入，返回清理函数：

```go
inj, cleanup, err := inject.InitInjector(ctx)
defer cleanup()
```

### 数据库操作

支持读写分离、事务管理、分页查询：

```go
// 获取 DB 实例
db := dbx.GetDB(ctx, injector.DB)

// 分页查询
result, err := dbx.WrapPageQuery(ctx, db.Model(&User{}), pp, dbx.QueryOptions{}, &users)

// 事务
err := trans.Exec(ctx, func(ctx context.Context) error {
    db := dbx.GetDB(ctx, trans.DB)
    // 执行数据库操作
    return nil
})
```

#### 排序：两条通道

`QueryOptions` 有两条排序通道，按「字段值从哪来」选：

| 通道 | 适用场景 | 行为 |
|---|---|---|
| `OrderFields` | 字段名可能来自 HTTP query | 标识符白名单 + 方言引号；**不合法项被静默丢弃** |
| `RawOrderFields` | 表达式由服务端代码构造 | 原样拼接，不加引号、不做白名单 |

```go
// ① 普通列排序：自动加方言引号（MySQL/SQLite 反引号、PG 双引号），
//    保留字列名（order / key）也能正常工作。
opts := dbx.QueryOptions{
    OrderFields: dbx.OrderByParams{
        {Field: "created_at", Direction: dbx.DESC},
        {Field: "id", Direction: dbx.ASC},
    },
}

// ② 表达式排序（CASE / 函数 / NULLS LAST）：走 RawOrderFields。
//    典型用法：先按表达式分层，再在同层内按普通列细分。
opts := dbx.QueryOptions{
    RawOrderFields: dbx.RawOrderFields{
        {
            Expr:      "CASE status WHEN 'enabled' THEN 0 WHEN 'auto_disabled' THEN 1 ELSE 2 END",
            Direction: dbx.ASC,
        },
    },
    OrderFields: dbx.OrderByParams{
        {Field: "priority", Direction: dbx.DESC},
    },
}
// 生成：ORDER BY CASE ... END ASC, `priority` DESC
```

> **⚠️ `RawOrderFields.Expr` 禁止直接接收 HTTP query 参数。**
> 它不做任何校验或转义，防注入责任完全在调用方。用户可控的排序需求请用
> `OrderFields`；若需要把用户输入映射成表达式，务必走服务端白名单
> （如 `switch orderBy { case "status": expr = "CASE ..." }`）。

`OrderFields` 静默丢弃非法项是刻意取舍（排序不了不该让查询失败），但丢弃
是可以被观测到的：

```go
// handler 层：把非法 ?order_by= 判成 400，而不是静默忽略
if err := params.OrderFields.Validate(); err != nil {
    return web.ResError(c, errors.BadRequest("", "%s", err))
}

// 中间层：只想告警、不想中断请求
if dropped := params.OrderFields.DroppedFields(); len(dropped) > 0 {
    logger.From(ctx).Warn("orderBy fields dropped", zap.Strings("fields", dropped))
}
```

### 认证中间件

基于 JWT 的认证，支持多种存储后端：

```go
// 配置认证中间件
auth := middleware.Auth(inj.Auth, middleware.AuthConfig{
    SkippedPathPrefixes: []string{"/health", "/api/v1/login"},
})
```

### 统一响应格式

```go
// 成功响应
web.ResOk(c)
web.ResSuccess(c, data)
web.ResPage(c, list, total)

// 错误响应
web.ResError(c, errors.BadRequest("无效参数"))
```

### CRUD 功能

开箱即用的 CRUD 实现：

```go
type UserModule struct {
    crud.BaseCRUD
}

func (m *UserModule) Init(ctx context.Context) error {
    m.SetModel(&User{})
    return nil
}
```

## 典型应用

```go
package main

import (
    "context"
    "reflect"

    "github.com/puras/mog/command"
    "github.com/puras/mog/inject"
    "github.com/puras/mog/jwtx"
    "github.com/puras/mog/module"
    "github.com/puras/mog/server"
    "github.com/puras/mog/web"
    "github.com/urfave/cli/v2"
)

type App struct {
    inj *inject.Injector
}

func (a *App) Init(ctx context.Context) error {
    inj, cleanup, err := inject.InitInjector(ctx)
    if err != nil {
        return err
    }
    a.inj = inj
    return nil
}

func (a *App) GetInjector(ctx context.Context) *inject.Injector {
    return a.inj
}

func (a *App) RegistryRoutes(ctx context.Context, e *gin.Engine) error {
    return module.RegistryRoutes(ctx, e, reflect.ValueOf(*a))
}

func (a *App) ParseCurrentUser(c *gin.Context) (string, error) {
    token := web.GetToken(c)
    return a.inj.Auth.ParseSubject(c.Request.Context(), token)
}

func main() {
    app := cli.NewApp()
    app.Name = "myapp"
    app.Commands = []*cli.Command{
        command.StartCmd(&App{}),
        command.StopCmd(),
    }
    app.Run(os.Args)
}
```

## 依赖管理

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./path/to/package
```

## 版本

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本变更历史。

## 许可证

[MIT License](LICENSE)
