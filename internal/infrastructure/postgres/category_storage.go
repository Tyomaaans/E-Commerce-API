package postgres

import (
	"time"

	"gorm.io/gorm"
)

type CategoryStorage struct {
	ID        string           `gorm:"type:uuid;primaryKey"`
	Name      string           `gorm:"type:varchar(100);not null"`
	Slug      string           `gorm:"type:varchar(120);uniqueIndex;not null"`
	Image     *string          `gorm:"type:varchar(255)"`
	ParentID  *string          `gorm:"type:uuid;index"`
	CreatedAt time.Time        `gorm:"not null"`
	UpdatedAt time.Time        `gorm:"not null"`
	DeletedAt gorm.DeletedAt   `gorm:"index"`

	Parent    *CategoryStorage `gorm:"foreignKey:ParentID"`
	Products  []ProductStorage `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}