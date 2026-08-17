package postgres

import (
	"time"

	"gorm.io/gorm"
)

type ProductStorage struct {
	ID            string                 `gorm:"type:uuid;primaryKey"`
	StoreID       string                 `gorm:"type:uuid;not null;uniqueIndex:idx_store_slug"`
	CategoryID    string                 `gorm:"type:uuid;not null;index"`
	Name          string                 `gorm:"type:varchar(255);not null"`
	Slug          string                 `gorm:"type:varchar(255);not null;uniqueIndex:idx_store_slug"`
	Description   string                 `gorm:"type:text;not null"`
	Price         float64                `gorm:"not null"`
	Stock         int                    `gorm:"not null;default:0;index:idx_stock_reserved,priority:1;check:chk_stock_valid, stock >= 0"`
	ReservedStock int                 	 `gorm:"not null;default:0;index:idx_stock_reserved,priority:2;check:chk_reserved_valid, reserved_stock >= 0 AND reserved_stock <= stock"` 
	WeightGram    int                    `gorm:"not null"`
	IsActive      bool                   `gorm:"not null;default:true"`
	CreatedAt     time.Time              `gorm:"not null"`
	UpdatedAt     time.Time              `gorm:"not null"`
	DeletedAt     gorm.DeletedAt         `gorm:"index"`
    
	Images        []ProductImageStorage  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Reviews       []ProductReviewStorage `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

type ProductImageStorage struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	ProductID string         `gorm:"type:uuid;not null;index"`
	ImageURL  string         `gorm:"type:varchar(255);not null"`
	IsPrimary bool           `gorm:"not null;default:false"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ProductReviewStorage struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	UserID    string         `gorm:"type:uuid;not null;index"`
	ProductID string         `gorm:"type:uuid;not null;index"`
	Rating    int            `gorm:"type:smallint;not null"`
	Comment   *string        `gorm:"type:text"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}