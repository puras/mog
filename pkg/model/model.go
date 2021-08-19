package model

import "time"

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 11:07
 * @desc
 */
type BaseModel struct {
	ID        string    `json:"id" gorm:"primary_key;unique_index;size:64"`
	Deleted   bool      `json:"-"'`
	CreatedBy string    `json:"createdBy" gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedBy string    `json:"updatedBy" gorm:"column:updated_by"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}
