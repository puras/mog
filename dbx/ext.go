package dbx

import (
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
