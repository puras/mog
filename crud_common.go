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
