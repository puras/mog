package model

import "mooko.net/mog/pkg/model"

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 08:58
 * @desc
 */

type Library struct {
	Name        string `json:"name" gorm:"size:255"`
	Description string `json:"description" gorm:"size:1024"`
	model.BaseModel
}

func (Library) TableName() string {
	return "ku_library_tab"
}
