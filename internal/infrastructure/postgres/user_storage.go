package postgres

import (
	"time"

	"gorm.io/gorm"

	"E-COMMERCE-API/internal/domain"
)

type UserStorage struct {
	ID                 string                 `gorm:"type:uuid;primaryKey"`
	Name               *string                `gorm:"type:varchar(255)"`
	Email              string                 `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username           *string                `gorm:"type:varchar(50);uniqueIndex"`
	Phone              *string                `gorm:"type:varchar(50);uniqueIndex"`
	Role               domain.Role            `gorm:"type:varchar(50);not null"`
	Password           string                 `gorm:"type:varchar(255);not null"`
	Avatar             *string                `gorm:"type:varchar(100)"`
	EmailVerifiedAt    *time.Time             `gorm:"index"`
	ProfileCompletedAt *time.Time             `gorm:"index"`
	RememberMe         bool                   `gorm:"not null;default:false"`
	CreatedAt          time.Time              `gorm:"not null"`
	UpdatedAt          time.Time              `gorm:"not null"`
	DeletedAt          gorm.DeletedAt         `gorm:"index"`

	Providers          []UserProviderStorage  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Addresses          []UserAddressStorage   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserProviderStorage struct {
    ID             string          `gorm:"type:uuid;primaryKey"`
    UserID         string          `gorm:"type:uuid;not null;uniqueIndex:idx_user_provider,priority:1"`
    Provider       domain.Provider `gorm:"type:varchar(50);not null;uniqueIndex:idx_user_provider,priority:2;uniqueIndex:idx_provider_account,priority:1"`
    ProviderUserID *string         `gorm:"type:varchar(255);uniqueIndex:idx_provider_account,priority:2"`
    Password       *string         `gorm:"type:varchar(255)"`
    CreatedAt      time.Time       `gorm:"not null"`
    UpdatedAt      time.Time       `gorm:"not null"`
    DeletedAt      gorm.DeletedAt  `gorm:"index"`
}

type UserAddressStorage struct {
    ID           string         `gorm:"type:uuid;primaryKey"`
    UserID       string         `gorm:"type:uuid;not null;index;uniqueIndex:idx_one_primary_per_user,priority:1"`
    Label        string         `gorm:"type:varchar(50);not null"`
    Recipient    string         `gorm:"type:varchar(100);not null"`
    Phone        string         `gorm:"type:varchar(50);not null"`
    AddressLine1 string         `gorm:"type:text;not null"`
    AddressLine2 *string        `gorm:"type:text"`
    City         string         `gorm:"type:varchar(100);not null"`
    State        string         `gorm:"type:varchar(100);not null"`
    PostalCode   string         `gorm:"type:varchar(20);not null"`
    Country      string         `gorm:"type:varchar(100);not null;default:'Indonesia'"`
    IsPrimary    *bool          `gorm:"not null;default:true;uniqueIndex:idx_one_primary_per_user,priority:2,where:is_primary = true"`
    CreatedAt    time.Time      `gorm:"not null"`
    UpdatedAt    time.Time      `gorm:"not null"`
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}	