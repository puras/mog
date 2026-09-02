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

func WhereLike(db *gorm.DB, field string, value string) {
	db.Where(field+"=?", LikeParameter(value))
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
