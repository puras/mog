# 变更日志

本文档记录 MoG 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added
- 项目文档（CLAUDE.md）

### Changed
- 重构依赖注入为手动实现（移除 Wire 依赖）
- 升级 Go 版本至 1.26

### Fixed
- 完成默认 CRUD 功能，Model 配合修改

## [0.1.4] - 2023-09-27

### Added
- 错误处理模块

### Changed
- 将可复用的内容迁移到 mog
- 重构依赖注入为手动实现

### Fixed
- 完成默认 CRUD 功能，Model 配合修改

## [0.1.3] - 2023-09-27

### Added
- MinIO 客户端配置增加 UseSSL 选项

### Changed
- 更新 MinIO 客户端函数名称，客户端安全默认设为 false

### Fixed
- 修正 MinIO 工具命令问题

## [0.1.2] - 2023-09-22

### Added
- 配置支持扩展配置（ext config）

## [0.1.1] - 2023-09-12

### Added
- 数据库配置增加开关控制

## [0.1.0] - 2023-09-08

### Added
- 初始版本发布
- 基于 Gin 的 Web 框架
- GORM 数据库支持（MySQL、PostgreSQL、SQLite）
- JWT 认证中间件
- 配置管理（TOML 格式）
- 日志系统（基于 zap）
- 错误处理
- MinIO 对象存储支持
- 邮件发送功能
- 缓存支持（内存、Redis）
- 读写分离支持

---

[Unreleased]: https://github.com/puras/mog/compare/v0.1.4...HEAD
[0.1.4]: https://github.com/puras/mog/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/puras/mog/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/puras/mog/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/puras/mog/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/puras/mog/releases/tag/v0.1.0
