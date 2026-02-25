package crud

import (
	"context"
	"time"

	"github.com/puras/mog/dbx"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/model"
	"gorm.io/gorm"
)

type ICrudRepo[T model.IModel] interface {
	GetModelDB(ctx context.Context) *gorm.DB
	Query(ctx context.Context, params QueryParams, pageParams dbx.PaginationParam, opts ...dbx.QueryOptions) (*dbx.PaginationResult, error)
	Get(ctx context.Context, id string, opts ...dbx.QueryOptions) (*T, error)
	Create(ctx context.Context, item *T) error
	Update(ctx context.Context, id string, item *T) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type FillQueryParametersFunc func(ctx context.Context, db *gorm.DB, params QueryParams)

type CrudRepo[T model.IModel] struct {
	DB                      *gorm.DB
	FillQueryParametersFunc FillQueryParametersFunc
}

func NewRepo[T model.IModel](db *gorm.DB, fillQueryParametersFunc FillQueryParametersFunc) *CrudRepo[T] {
	return &CrudRepo[T]{db, fillQueryParametersFunc}
}

func (self *CrudRepo[T]) GetModelDB(ctx context.Context) *gorm.DB {
	return dbx.GetDB(ctx, self.DB).Model(new(T))
}

func (self *CrudRepo[T]) Query(ctx context.Context, params QueryParams, pageParams dbx.PaginationParam, opts ...dbx.QueryOptions) (*dbx.PaginationResult, error) {
	var opt dbx.QueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := self.GetModelDB(ctx)

	if self.FillQueryParametersFunc != nil {
		self.FillQueryParametersFunc(ctx, db, params)
	}

	dbx.NotDeleted(db)

	var list []*T
	pr, err := dbx.WrapPageQuery(ctx, db, pageParams, opt, &list)
	return dbx.WrapPaginationResult(pr, list, err)
}

func (self *CrudRepo[T]) Get(ctx context.Context, id string, opts ...dbx.QueryOptions) (*T, error) {
	var opt dbx.QueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(T)
	db := self.GetModelDB(ctx)
	dbx.Where(db, "id", id)
	dbx.NotDeleted(db)
	ok, err := dbx.FindOne(ctx, db, opt, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (self *CrudRepo[T]) Create(ctx context.Context, item *T) error {
	ret := self.GetModelDB(ctx).Create(item)
	return errors.WithStack(ret.Error)
}

func (self *CrudRepo[T]) Update(ctx context.Context, id string, item *T) error {
	db := self.GetModelDB(ctx)
	dbx.Where(db, "id", id)
	ret := db.Select("*").Omit("created_at", "deleted", "deleted_at").Updates(item)
	return errors.WithStack(ret.Error)
}

// Delete 逻辑删除
func (self *CrudRepo[T]) Delete(ctx context.Context, id string) error {
	db := self.GetModelDB(ctx)
	dbx.Where(db, "id", id)
	ret := db.Omit("updated_at").Updates(map[string]interface{}{"deleted": true, "deleted_at": time.Now()})
	return errors.WithStack(ret.Error)
}

// HardDelete 物理删除
func (self *CrudRepo[T]) HardDelete(ctx context.Context, id string) error {
	db := self.GetModelDB(ctx)
	dbx.Where(db, "id", id)
	ret := db.Delete(new(T))
	return errors.WithStack(ret.Error)
}
