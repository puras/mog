package dbx

import (
	"context"
	"github.com/puras/mog/contextx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaginationResult struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

type PaginationParam struct {
	Pagination bool `form:"-" default:"true"`
	OnlyCount  bool `form:"-"`
	PageNum    int  `form:"page_num"`
	PageSize   int  `form:"page_size" binding:"max=100"`
}

type QueryOptions struct {
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
}

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam

func (p OrderByParams) ToSQL() string {
	if len(p) == 0 {
		return ""
	}

	var sql string
	for _, v := range p {
		sql += v.Field + " " + string(v.Direction) + ","
	}
	return sql[:len(sql)-1]
}

func (t *Trans) Exec(ctx context.Context, fn TransFunc) error {
	if _, ok := contextx.FromTrans(ctx); ok {
		return fn(ctx)
	}
	return t.DB.Transaction(func(db *gorm.DB) error {
		return fn(contextx.NewTrans(ctx, db))
	})
}

func GetDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	db := defDB

	if tdb, ok := contextx.FromTrans(ctx); ok {
		db = tdb
	}
	if contextx.FromRowLock(ctx) {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return db.WithContext(ctx)
}

func wrapQueryOptions(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if len(opts.SelectFields) > 0 {
		db = db.Select(opts.SelectFields)
	}
	if len(opts.OmitFields) > 0 {
		db = db.Omit(opts.OmitFields...)
	}
	if len(opts.OrderFields) > 0 {
		db = db.Order(opts.OrderFields.ToSQL())
	}
	return db
}

func WrapPageQuery(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out any) (*PaginationResult, error) {
	if pp.OnlyCount {
		var count int64
		err := db.Count(&count).Error
		if err != nil {
			return nil, err
		}
		return &PaginationResult{Total: count}, nil
	} else if !pp.Pagination {
		pageSize := pp.PageSize
		if pageSize > 0 {
			db = db.Limit(pageSize)
		}
		db = wrapQueryOptions(db, opts)
		err := db.Find(out).Error
		return nil, err
	}
	total, err := FindPage(ctx, db, pp, opts, out)
	if err != nil {
		return nil, err
	}
	return &PaginationResult{
		Items:    out,
		Total:    total,
		PageNum:  pp.PageNum,
		PageSize: pp.PageSize,
	}, nil
}

func FindPage(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out any) (int64, error) {
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	} else if count == 0 {
		return count, nil
	}

	pageNum, pageSize := pp.PageNum, pp.PageSize
	if pageNum > 0 && pageSize > 0 {
		db = db.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	} else if pageSize > 0 {
		db = db.Limit(pageSize)
	}

	db = wrapQueryOptions(db, opts)
	err = db.Find(out).Error
	return count, err
}

func FindOne(ctx context.Context, db *gorm.DB, opts QueryOptions, out any) (bool, error) {
	db = wrapQueryOptions(db, opts)
	result := db.Limit(1).Scan(out)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func Exists(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	result := db.Count(&count)
	if err := result.Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
