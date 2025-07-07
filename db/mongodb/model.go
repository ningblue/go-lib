package mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" `
	CreatedAt time.Time          `bson:"created_at" json:"created_at" `
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at" `
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty" `
}
