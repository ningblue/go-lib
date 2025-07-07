package gorm

import (
	"lib/db/objectID"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type BaseObjectIDModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	NumberID  string         `gorm:"type:varchar(128);not null;uniqueIndex" json:"number_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (base *BaseObjectIDModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.NumberID == "" {
		base.NumberID = objectID.NewObjectID().Hex()
	}
	return
}
