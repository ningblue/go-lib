package gorm

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

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

// UUIDBaseModel UUID 主键基座（主键由数据库 gen_random_uuid() 生成，应用侧不造 ID），无软删。
// 适用于凭据/配置/只增改等不需要软删语义的表。
type UUIDBaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UUIDSoftDeleteModel UUID 主键 + 软删基座。
// 配合部分唯一索引（UNIQUE ... WHERE deleted_at IS NULL）实现"软删后同名可重建"语义。
type UUIDSoftDeleteModel struct {
	UUIDBaseModel
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
