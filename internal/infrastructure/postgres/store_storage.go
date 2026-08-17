package postgres

import (
	"time"

	"gorm.io/gorm"
)

type StoreStorage struct {
	ID          string           `gorm:"type:uuid;primaryKey"`
	UserID      string           `gorm:"type:uuid;not null;uniqueIndex"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Slug        string           `gorm:"type:varchar(120);uniqueIndex;not null"`
	Description *string          `gorm:"type:text"`
	Logo        *string          `gorm:"type:varchar(255)"`
	IsActive    bool             `gorm:"not null;default:true"`
	CreatedAt   time.Time        `gorm:"not null"`
	UpdatedAt   time.Time        `gorm:"not null"`
	DeletedAt   gorm.DeletedAt   `gorm:"index"`

	User        *UserStorage     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Products    []ProductStorage `gorm:"foreignKey:StoreID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}