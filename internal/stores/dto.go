package stores

import (
	"E-COMMERCE-API/internal/products"
	"E-COMMERCE-API/internal/users"
)

// ==== Request ====

type RegisterStoreRequest struct {
	UserID      string  `json:"-"`
	Name        string  `json:"name"        validate:"required,alphaspaceunicode"`
	Slug        string  `json:"-"`
	Description *string `json:"description" validate:"omitempty"`
	Logo        *string `json:"logo"        validate:"omitempty"`
	IsActive    bool    `json:"-"`
}

type UpdateStoreRequest struct {
	ID          string  `json:"-"`
	UserID      string  `json:"-"`
	Name        *string `json:"name"        validate:"omitempty,alphaspaceunicode"`
	Slug        *string `json:"-"`
	Description *string `json:"description" validate:"omitempty"`
	Logo        *string `json:"logo"        validate:"omitempty"`
}

// ==== Response ====

type StoreResponse struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id,omitempty"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description *string            `json:"description,omitempty"`
	Logo        *string            `json:"logo,omitempty"`
	IsActive    bool               `json:"is_active"`

	User        *users.UserResponse `json:"user,omitempty"`
	Products    []products.ProductResponse `json:"product,omitempty"`
}