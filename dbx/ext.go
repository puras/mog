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

// WhereIn 在 field 上加 IN 匹配。
//
// 生成 `field IN ?` + values（gorm 会把 slice 展开为 `(?, ?, ...)`）。
// values 为空时直接跳过，避免某些方言（MySQL/PostgreSQL）下 IN () 语法错误；
// 与 WhereLike 一致，不为单一方言提供分支 helper。
//
// 相比 []any 版本，使用泛型 T 让调用方可以直接传 []int64 / []string 等
// typed slice（无 boxing 拷贝），编译期即拒掉异构元素——SQL IN 天然同构，
// 把约束上提到类型层比留到运行时更合算。
func WhereIn[T any](db *gorm.DB, field string, values []T) {
	if len(values) == 0 {
		return
	}
	db.Where(field+" IN ?", values)
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
