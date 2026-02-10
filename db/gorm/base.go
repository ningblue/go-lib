package gorm

import (
	"encoding/json"
	"time"

	"github.com/ningblue/go-lib/db/objectID"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MarshalJSON 自定义 JSON 序列化，返回 ISO 8601 格式带时区的时间
func (m BaseModel) MarshalJSON() ([]byte, error) {
	type Alias BaseModel
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(&m),
	})
}

type BaseObjectIDModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	NumberID  string         `gorm:"type:varchar(128);not null;uniqueIndex" json:"number_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MarshalJSON 自定义 JSON 序列化，返回 ISO 8601 格式带时区的时间
func (m BaseObjectIDModel) MarshalJSON() ([]byte, error) {
	type Alias BaseObjectIDModel
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
		Alias:     (*Alias)(&m),
	})
}

func (base *BaseObjectIDModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.NumberID == "" {
		base.NumberID = objectID.NewObjectID().Hex()
	}
	return
}
