# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

MoG (Mooko Go) 是一个 Go Web 基础框架，基于 Gin、Gorm、Wire 等开源项目构建，用于快速开发 Web 应用。

## 常用命令

### 构建和运行
```bash
# 构建项目
go build

# 运行服务器（使用默认配置文件 conf/config.toml）
./app start

# 指定配置文件运行
./app start --conf conf/config.toml

# 以守护进程模式运行
./app start --daemon

# 停止服务器
./app stop
```

### 依赖管理
```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./path/to/package
```

## 项目架构

### 核心模块

#### 1. 配置管理 (`config/`)
- 使用 TOML 格式配置文件
- 配置通过 `config.MustLoad("conf/config.toml")` 加载
- 主要配置项：
  - `General`: HTTP 服务配置（端口、超时等）
  - `Storage`: 存储配置（缓存、数据库）
  - `Logger`: 日志配置
  - `Middleware`: 中间件配置（CORS、认证、限流等）

#### 2. 依赖注入 (`inject/`)
- 使用手动依赖注入
- `Injector` 结构体包含：`*gorm.DB`、`jwtx.Auth` 和 `*dbx.Trans`
- 通过 `inject.InitInjector()` 初始化注入器
- 返回清理函数用于资源释放

#### 3. 服务器启动 (`server/`)
- `ServerParam` 接口定义应用启动所需的参数：
  - `Init()`: 初始化回调
  - `GetInjector()`: 获取依赖注入器
  - `RegistryRoutes()`: 注册路由
  - `ParseCurrentUser()`: 解析当前用户（用于认证）
- 服务器通过 `server.Run()` 启动，支持优雅关闭
- 默认中间件：Recovery、Trace、Logger、Auth

#### 4. 数据库 (`dbx/`)
- 支持 MySQL、PostgreSQL、SQLite3
- 支持读写分离（通过 dbresolver 插件）
- 事务管理：通过 `Trans.Exec()` 执行事务
- 上下文集成：`GetDB(ctx, defDB)` 获取带上下文的 DB 实例
- 分页查询：`WrapPageQuery()` 封装分页逻辑
- 模型层提供 `Model`、`DefaultModel`、`BaseModel`、`TenantModel` 基类

#### 5. 认证 (`jwtx/`)
- 基于 JWT 的认证
- 支持多种存储后端：Badger、Redis、Memory
- 默认签名方法：HS512
- Token 存储在缓存中，支持撤销

#### 6. 中间件 (`middleware/`)
- `Recovery`: 恢复 panic
- `Trace`: 请求追踪（X-Request-Id、X-Trace-Id）
- `Logger`: 请求日志
- `Auth`: 认证中间件，支持跳过特定路径前缀

#### 7. Web 响应 (`web/`)
- 统一响应格式：`ResponseResult{Code, Data, Message}`
- 成功响应：`ResOk()`, `ResSuccess()`, `ResPage()`
- 错误响应：`ResError()` 自动记录日志
- 请求解析：`ParseJSON()`, `ParseQuery()`, `ParseForm()`

#### 8. 错误处理 (`errors/`)
- 自定义 `Error` 类型：包含 `Id`、`Code`、`Detail`、`Status`
- 预定义错误：`BadRequest`、`Unauthorized`、`Forbidden`、`NotFound` 等
- 错误堆栈：支持 `WithStack`、`Wrap`、`Wrapf`

#### 9. 模块系统 (`module/`)
- `Module` 接口：`Init()` 和 `RegistryRoutes()`
- 通过反射自动初始化和注册模块字段

#### 10. 日志 (`logger/`)
- 基于 zap 结构化日志
- 上下文支持：trace_id、user_id、tag
- 预定义标签：TagKeyMain、TagKeyRequest、TagKeySystem 等

### 典型应用结构

```go
// 1. 定义 ServerParam 实现
type App struct {
    inj *inject.Injector
}

func (a *App) Init(ctx context.Context) error {
    // 初始化模块
    return module.Init(ctx, reflect.ValueOf(*a))
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

// 2. 定义业务模块
type UserModule struct {
    // ...
}

func (m *UserModule) Init(ctx context.Context) error {
    // 初始化逻辑
    return nil
}

func (m *UserModule) RegistryRoutes(ctx context.Context, e *gin.Engine) error {
    v1 := e.Group("/api/v1")
    users := v1.Group("/users")
    {
        users.GET("", m.list)
        users.POST("", m.create)
    }
    return nil
}

// 3. 在 main.go 中使用
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

### 数据库查询模式

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

### 配置文件示例

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

## 注意事项

1. **手动依赖注入**：使用 `inject.InitInjector(ctx)` 初始化依赖，记得在结束时调用清理函数
2. **模型继承**：业务模型应继承 `model.Model`、`model.DefaultModel` 或 `model.BaseModel`
3. **上下文传递**：数据库操作需要传递 context，使用 `dbx.GetDB(ctx, defDB)`
4. **错误处理**：使用 `errors.*` 预定义错误函数，错误会自动转换为 JSON 响应
5. **日志记录**：使用 `logger.Context(ctx)` 获取带上下文的 logger
