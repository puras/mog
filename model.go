package mog

import (
	"time"
)

/**
 * @project kuko
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-19 11:07
 * @desc
 */
type Model struct {
	ID        string `json:"id" gorm:"primary_key;unique_index;size:64"`
	CreatedAt Time   `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt Time   `json:"updatedAt" gorm:"column:updated_at"`
}

type DefaultModel struct {
	Model
	Deleted bool `json:"-"`
}

type BaseModel struct {
	DefaultModel
	CreatedBy string `json:"createdBy" gorm:"column:created_by"`
	UpdatedBy string `json:"updatedBy" gorm:"column:updated_by"`
}

func (b *Model) DefaultCreated() {
	var now = Time(time.Now())
	b.CreatedAt = now
	b.UpdatedAt = now
	b.ID = GenShortUUID()
}

func (b *Model) DefaultUpdated() {
	b.UpdatedAt = Time(time.Now())
}
