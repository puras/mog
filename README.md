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

Mooko Go WEB基础框架。基于Gin、Gorm、Wire等开源项目整理成基础工具，方便构建WEB工程。