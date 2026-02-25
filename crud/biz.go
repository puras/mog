package crud

import (
	"context"

	"github.com/puras/mog/dbx"
	"github.com/puras/mog/errors"
	"github.com/puras/mog/model"
	"github.com/puras/mog/web"
)

type ICrudBiz[T model.IModel, V model.IForm[T]] interface {
	Query(ctx context.Context, params QueryParams) (*web.PageResult, error)
	Get(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, item V) (*T, error)
	Update(ctx context.Context, id string, item V) error
	Delete(ctx context.Context, id string) error
}

type UpdateQueryOptionsFunc func(ctx context.Context) (dbx.QueryOptions, error)

type CrudBiz[T model.IModel, V model.IForm[T]] struct {
	Trans                  *dbx.Trans
	Repo                   ICrudRepo[T]
	UpdateQueryOptionsFunc UpdateQueryOptionsFunc
}

func NewBiz[T model.IModel, V model.IForm[T]](
	trans *dbx.Trans,
	repo ICrudRepo[T],
	updateQueryOptionsFunc UpdateQueryOptionsFunc,
) *CrudBiz[T, V] {
	return &CrudBiz[T, V]{
		Trans:                  trans,
		Repo:                   repo,
		UpdateQueryOptionsFunc: updateQueryOptionsFunc,
	}
}

func (self *CrudBiz[T, V]) Query(ctx context.Context, params QueryParams) (*web.PageResult, error) {
	queryOptions := dbx.QueryOptions{
		OrderFields: []dbx.OrderByParam{
			{Field: "created_at", Direction: dbx.DESC},
		},
	}
	if self.UpdateQueryOptionsFunc != nil {
		opts, err := self.UpdateQueryOptionsFunc(ctx)
		if err != nil {
			return nil, err
		}
		queryOptions = opts
	}
	pageParams := params.GetPaginationParam()
	pageParams.Pagination = true
	ret, err := self.Repo.Query(ctx, params, pageParams, queryOptions)
	if err != nil {
		return nil, err
	}
	return web.FromPaginationResult(ret), nil
}

func (self *CrudBiz[T, V]) Get(ctx context.Context, id string) (*T, error) {
	item, err := self.Repo.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.NotFound(id, "用户不存在")
	}
	return item, nil
}

func (self *CrudBiz[T, V]) Create(ctx context.Context, form V) (*T, error) {
	item := new(T)
	if err := form.FillTo(item); err != nil {
		return nil, err
	}

	if defaulter, ok := any(item).(interface{ DefaultCreated() }); ok {
		defaulter.DefaultCreated()
	}

	err := self.Trans.Exec(ctx, func(ctx context.Context) error {
		return self.Repo.Create(ctx, item)
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (self *CrudBiz[T, V]) Update(ctx context.Context, id string, form V) error {
	item, err := self.Repo.Get(ctx, id)
	if err != nil {
		return err
	} else if item == nil {
		return errors.NotFound(id, "数据不存在")
	}

	if err := form.FillTo(item); err != nil {
		return err
	}

	if defaulter, ok := any(item).(interface{ DefaultUpdated() }); ok {
		defaulter.DefaultUpdated()
	}

	return self.Trans.Exec(ctx, func(ctx context.Context) error {
		return self.Repo.Update(ctx, id, item)
	})
}

func (self *CrudBiz[T, V]) Delete(ctx context.Context, id string) error {
	item, err := self.Repo.Get(ctx, id)
	if err != nil {
		return err
	} else if item == nil {
		return errors.NotFound("", "数据不存在")
	}

	return self.Trans.Exec(ctx, func(ctx context.Context) error {
		return self.Repo.Delete(ctx, id)
	})
}
