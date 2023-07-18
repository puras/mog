package model

import (
	"github.com/puras/mog/utils"
	"time"
)

type Model struct {
	ID        string    `json:"id" gorm:"primary_key;unique_index;size:64"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
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

func (m *Model) DefaultCreated() {
	var now = time.Now()
	m.ID = utils.NewID()
	m.CreatedAt = now
	m.UpdatedAt = now
}

func (m *Model) DefaultUpdated() {
	m.UpdatedAt = time.Now()
}
