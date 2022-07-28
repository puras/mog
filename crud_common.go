package mog

import (
	"gorm.io/gorm"
)

/**
 * @project momo-backend
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-09-01 16:36
 * @desc
 */

func NoDelete(db **gorm.DB) {
	(*db).Where("deleted", false)
}

func DoDelete(db **gorm.DB) {
	(*db).Update("deleted", true)
}

type CRUD interface {
	DoDelete(db **gorm.DB) CRUD
	NoDelete(db **gorm.DB) CRUD
}

func NewCRUD() CRUD {
	return &crud{}
}

func Offset(pageNum, pageSize int) int {
	return (pageNum - 1) * pageSize
}

type crud struct{}

func (self crud) DoDelete(db **gorm.DB) CRUD {
	DoDelete(db)
	return self
}

func (self crud) NoDelete(db **gorm.DB) CRUD {
	NoDelete(db)
	return self
}
