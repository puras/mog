package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/puras/mog/model"
	"github.com/puras/mog/web"
)

type ICrudApi[T model.IModel, D QueryParams, F model.IForm[T]] interface {
	Query(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type CrudApi[T model.IModel, D QueryParams, F model.IForm[T]] struct {
	Biz ICrudBiz[T, F]
}

func NewApi[T model.IModel, D QueryParams, F model.IForm[T]](
	biz ICrudBiz[T, F],
) *CrudApi[T, D, F] {
	return &CrudApi[T, D, F]{
		Biz: biz,
	}
}

func (self *CrudApi[T, D, F]) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params D
	if err := web.ParseQuery(c, &params); err != nil {
		web.ResError(c, err)
		return
	}

	ret, err := self.Biz.Query(ctx, params)
	if err != nil {
		web.ResError(c, err)
		return
	}
	web.ResPage(c, ret)
}

func (self *CrudApi[T, D, F]) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param(web.PARAM_ID)
	item, err := self.Biz.Get(ctx, id)
	if err != nil {
		web.ResError(c, err)
		return
	}
	web.ResSuccess(c, item)
}

func (self *CrudApi[T, D, F]) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(F)
	if err := web.ParseJSON(c, item); err != nil {
		web.ResError(c, err)
		return
	} else if err := (*item).Validate(); err != nil {
		web.ResError(c, err)
		return
	}

	ret, err := self.Biz.Create(ctx, *item)
	if err != nil {
		web.ResError(c, err)
		return
	}
	web.ResSuccess(c, ret)
}

func (self *CrudApi[T, D, F]) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param(web.PARAM_ID)
	item := new(F)
	if err := web.ParseJSON(c, item); err != nil {
		web.ResError(c, err)
		return
	} else if err := (*item).Validate(); err != nil {
		web.ResError(c, err)
		return
	}

	err := self.Biz.Update(ctx, id, *item)
	if err != nil {
		web.ResError(c, err)
		return
	}
	web.ResOk(c)
}

func (self *CrudApi[T, D, F]) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param(web.PARAM_ID)
	err := self.Biz.Delete(ctx, id)
	if err != nil {
		web.ResError(c, err)
		return
	}
	web.ResOk(c)
}
