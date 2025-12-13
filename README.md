MoG
----

Mooko Go WEB基础框架，基于Gin、Gorm、Wire等开源项目整理成基础工具集，旨在简化Go Web应用的开发流程，提供一套完整的解决方案，方便开发者快速构建高质量的WEB工程。

## 核心特性

- **基于Gin的Web框架**：高性能的HTTP路由和中间件支持
- **基于GORM的数据库操作**：支持多种数据库（MySQL、PostgreSQL、SQLite），提供统一的ORM接口
- **基于Wire的依赖注入**：简化组件间的依赖管理，提高代码的可测试性和可维护性
- **完善的中间件支持**：内置认证、日志、恢复、追踪等常用中间件
- **JWT认证机制**：提供安全可靠的用户认证解决方案
- **多种缓存支持**：集成Redis和BadgerDB缓存，支持灵活的缓存策略
- **日志管理**：基于Zap的高性能日志库，支持多种输出格式和级别
- **对象存储支持**：集成MinIO，方便管理和访问对象存储
- **命令行工具**：提供便捷的命令行操作，支持服务的启动、停止等
- **邮件发送功能**：支持SMTP邮件发送，方便实现通知功能

## 目录结构

```
├── cachex/         # 缓存管理（Redis、Badger）
├── command/        # 命令行工具
├── config/         # 配置管理
├── contextx/       # 上下文扩展
├── crypto/         # 加密工具
├── dbx/            # 数据库操作（基于GORM）
├── errors/         # 错误处理
├── inject/         # 依赖注入（基于Wire）
├── jwtx/           # JWT认证
├── logger/         # 日志管理
├── mail/           # 邮件发送
├── middleware/     # 中间件
├── model/          # 基础模型
├── module/         # 模块管理
├── oss/            # 对象存储（MinIO）
├── server/         # 服务器管理
├── utils/          # 工具函数
├── web/            # Web框架扩展（基于Gin）
├── .gitignore
├── LICENSE
├── README.md
├── go.mod
└── go.sum
```

## 快速开始

### 安装

```bash
go get -u github.com/puras/mog
```

### 简单示例

```go
package main

import (
    "github.com/puras/mog/server"
    "github.com/puras/mog/web"
)

func main() {
    // 创建Web服务器
    srv := server.NewServer()
    
    // 获取Gin引擎
    ginx := web.NewGinX()
    
    // 注册路由
    ginx.GET("/", func(c *web.Context) {
        c.JSON(200, web.H{
            "message": "Hello, MoG!",
        })
    })
    
    // 设置Web服务器的处理器
    srv.SetHandler(ginx)
    
    // 启动服务器
    srv.Start()
}
```

## 依赖说明

| 依赖项 | 版本 | 用途 |
|-------|------|------|
| github.com/gin-gonic/gin | v1.9.1 | Web框架 |
| gorm.io/gorm | v1.25.2 | ORM框架 |
| github.com/google/wire | v0.5.0 | 依赖注入 |
| github.com/golang-jwt/jwt | v3.2.2+incompatible | JWT认证 |
| github.com/redis/go-redis/v9 | v9.0.5 | Redis客户端 |
| go.uber.org/zap | v1.24.0 | 日志库 |
| github.com/dgraph-io/badger/v4 | v4.1.0 | KV存储 |
| github.com/minio/minio-go/v7 | v7.0.61 | 对象存储 |
| github.com/urfave/cli/v2 | v2.25.7 | 命令行工具 |
| gopkg.in/gomail.v2 | v2.0.0-20160411212932-81ebce5c23df | 邮件发送 |

## 使用示例

### 配置管理

```go
import "github.com/puras/mog/config"

// 加载配置
cfg := config.Load()

// 获取配置值
port := cfg.GetInt("server.port")
host := cfg.GetString("server.host")
```

### 数据库操作

```go
import (
    "github.com/puras/mog/dbx"
    "github.com/puras/mog/model"
)

// 定义模型
type User struct {
    model.Model
    Name  string
    Email string
}

// 初始化数据库
db := dbx.NewDB()

// 自动迁移模型
db.AutoMigrate(&User{})

// 创建用户
user := &User{Name: "test", Email: "test@example.com"}
db.Create(user)

// 查询用户
var foundUser User
db.First(&foundUser, user.ID)
```

### 中间件使用

```go
import (
    "github.com/puras/mog/middleware"
    "github.com/puras/mog/web"
)

ginx := web.NewGinX()

// 使用日志中间件
ginx.Use(middleware.Logger())

// 使用恢复中间件
ginx.Use(middleware.Recover())

// 使用认证中间件
ginx.Use(middleware.Auth())
```

### JWT认证

```go
import "github.com/puras/mog/jwtx"

// 创建JWT管理器
jwtManager := jwtx.NewJWT()

// 生成Token
token, err := jwtManager.GenerateToken("user123")

// 验证Token
claims, err := jwtManager.ParseToken(token)
```

### 日志记录

```go
import "github.com/puras/mog/logger"

// 获取日志实例
log := logger.GetLogger()

// 记录不同级别的日志
log.Debug("debug message")
log.Info("info message")
log.Warn("warn message")
log.Error("error message")
```

## 贡献指南

欢迎提交Issue和Pull Request！

## 许可证

本项目采用MIT许可证，详见[LICENSE](LICENSE)文件。