package dbx

import (
	"fmt"

	"github.com/puras/mog/errors"
	"gorm.io/gorm"
)

func WrapPaginationResult(pr *PaginationResult, list any, err error) (*PaginationResult, error) {
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if pr == nil {
		return &PaginationResult{
			Total: 0,
			Items: list,
		}, nil
	}

	return pr, nil
}

func LikeParameter(v string) string {
	return "%" + v + "%"
}

func NotDeleted(db *gorm.DB) {
	db.Where("deleted=false")
}

func Where(db *gorm.DB, field string, value any) {
	db.Where(field+"=?", value)
}

func WhereId(db *gorm.DB, value any) {
	db.Where("id=?", value)
}

// WhereLike 在 field 上加 LIKE 通配符匹配（跨方言大小写敏感性见下）。
//
// 生成 `field LIKE ?` + LikeParameter(value) → `LIKE '%value%'`。
// 大小写敏感性取决于方言：
//   - PostgreSQL：LIKE 默认大小写敏感（若需不敏感请用 raw SQL LOWER(field) LIKE LOWER(?)）
//   - MySQL / SQLite（ASCII）：LIKE 默认大小写不敏感
//
// 调用方如需 PG 上的大小写不敏感行为，请自行实现 LOWER 包装（mog 不为单一方言
// 提供方言分支 helper，避免 API 面被方言分歧污染）。
func WhereLike(db *gorm.DB, field string, value string) {
	db.Where(field+" like ?", LikeParameter(value))
}

// GetTableName 通过 gorm.Statement.Parse 获取 model 对应的表名（动态，非硬编码）。
// Parse 失败时回退到 NamingStrategy.TableName。
//
// 使用独立 Statement 实例避免跨调用状态污染，结果确定性与 NamingStrategy 一致。
func GetTableName(db *gorm.DB, model any) string {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return db.NamingStrategy.TableName(fmt.Sprintf("%T", model))
	}
	return stmt.Table
}
