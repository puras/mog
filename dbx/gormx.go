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

func (self PaginationParam) GetPaginationParam() PaginationParam {
	return self
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
	// WithContext 必须显式带上：GORM 的 Begin() 用 db.Statement.Context 作为事务 ctx，
	// 而 root DB 的 Statement.Context 是 context.Background()。漏掉它会让事务失去
	// deadline / cancel，调用方 contextx.NewTrans(ctx, ...) 拿到的 tx 也不认 ctx，
	// ctx 过期后语句照跑、事务无法自行收敛（连接带着未关闭的 tx 回到池里）。
	return t.DB.WithContext(ctx).Transaction(func(db *gorm.DB) error {
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
