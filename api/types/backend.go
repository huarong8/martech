package types

import (
	"time"

	"gorm.io/gorm"
)

// Backend 后台API
type Backend struct {
	ID          uint64         `gorm:"column:id;primaryKey"                   json:"id,omitempty"`
	CreatedAt   time.Time      `gorm:"column:created_at"                      json:"created_at,omitempty"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"                      json:"updated_at,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"column:delete_at;index"                 json:"-"`
	Path        string         `gorm:"column:path;index:idx_backend,unique"   json:"path,omitempty"`
	Method      string         `gorm:"column:method;index:idx_backend,unique" json:"method,omitempty"`
	Description string         `gorm:"column:description"                     json:"description,omitempty"`
	Group       string         `gorm:"column:group"                           json:"group,omitempty"`
	ObjectID    uint64         `gorm:"column:object_id"                       json:"object_id,omitempty"`
	Object      *Object        `gorm:"foreignKey:ObjectID"                    json:"object"`
}
