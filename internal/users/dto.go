package users

import (
	"time"

	"E-COMMERCE-API/internal/domain"
)

// ==== Request ====

type RegisterUserRequest struct {
	Email    string      `json:"email"    validate:"required,email"`
	Password string      `json:"password" validate:"required,password"`
	Role     domain.Role `json:"-"        validate:"required,oneof=seller buyer"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type LoginUserRequest struct {
	Email      string `json:"email"      validate:"required,email"`
	Password   string `json:"password"   validate:"required,password"`
	RememberMe bool   `json:"remember_me" validate:"omitempty"`
}

type CompleteProfileRequest struct {
	UserID   string  `json:"-"`
	Name     string  `json:"name"     validate:"required,alphaspaceunicode"`
	Username string  `json:"username" validate:"required,alphanum,min=4"`
	Phone    string  `json:"phone"    validate:"required,phone"`
	Avatar   *string `json:"avatar"   validate:"omitempty,url"`
}

type AddAddressRequest struct {
	UserID       string  `json:"-"              validate:"required"`
	Label        string  `json:"label"          validate:"required"`
	Recipient    string  `json:"receipent"      validate:"required"`
	Phone        string  `json:"phone"          validate:"required,phone"`
	AddressLine1 string	 `json:"address_line_1" validate:"required"`
	AddressLine2 *string `json:"address_line_2" validate:"omitempty"`
	City         string  `json:"city"           validate:"required"`
	State        string	 `json:"state"          validate:"required"`
	PostalCode   string  `json:"postal_code"    validate:"required"`
	Country      string  `json:"country"        validate:"required"`
	IsPrimary    *bool   `json:"is_primary"     validate:"omitempty"`
}

type UpdateUserRequest struct {
	ID                 string       `json:"-" uri:"id"`
	Name               *string      `json:"name"                 validate:"omitempty,alphaspaceunicode"`
	Email              *string      `json:"email"                validate:"omitempty,email"`
	Username           *string      `json:"username"             validate:"omitempty,alphanum,min=4"`
	Phone              *string      `json:"phone"                validate:"omitempty,phone"`
	Role               *domain.Role `json:"role"                 validate:"omitempty"`
	Avatar             *string      `json:"avatar"               validate:"omitempty,url"`
	EmailVerifiedAt    *time.Time   `json:"email_verified_at"    validate:"omitempty"`
	ProfileCompletedAt *time.Time   `json:"profile_completed_at" validate:"omitempty"`
	RememberMe         *bool        `json:"remember_me"          validate:"omitempty"`
}

type UpdateMySelfRequest struct {
	ID       string  `json:"-"`
	Name     *string `json:"name"     validate:"omitempty,alphaspaceunicode"`
	Username *string `json:"username" validate:"omitempty,alphanum,min=4"`
	Avatar   *string `json:"avatar"   validate:"omitempty,url"`
}

type UpdateAddressRequest struct {
	ID           string  `json:"-" uri:"id"`
	UserID       string  `json:"-"              validate:"required"`
	Label        *string `json:"label"          validate:"omitempty"`
	Recipient    *string `json:"receipent"      validate:"omitempty"`
	Phone        *string `json:"phone"          validate:"omitempty,phone"`
	AddressLine1 *string `json:"address_line_1" validate:"omitempty"`
	AddressLine2 *string `json:"address_line_2" validate:"omitempty"`
	City         *string `json:"city"           validate:"omitempty"`
	State        *string `json:"state"          validate:"omitempty"`
	PostalCode   *string `json:"postal_code"    validate:"omitempty"`
	Country      *string `json:"country"        validate:"omitempty"`
	IsPrimary    *bool   `json:"is_primary"     validate:"omitempty"`
}

type CreateUserProviderRequest struct {
	UserID         string          `json:"user_id"          validate:"required"`
	Provider       domain.Provider `json:"provider"         validate:"required"`
	ProviderUserID *string         `json:"provider_user_id" validate:"omitempty"`
	Password       *string         `json:"password"         validate:"omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

type ChangePassword struct {
	Email       string `json:"-"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password"`
}

// ==== Response ====

type UserResponse struct {
	ID                 string                 `json:"id"`
	Name               *string                `json:"name,omitempty"`
	Email              string                 `json:"email"`
	Username           *string                `json:"username,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	Role               string                 `json:"role"`
	Avatar             *string                `json:"avatar,omitempty"`
	EmailVerifiedAt    *string                `json:"email_verified_at,omitempty"`
	ProfileCompletedAt *string                `json:"profile_completed_at,omitempty"`
	RememberMe         bool                   `json:"remember_me"`
	Providers          []UserProviderResponse `json:"providers,omitempty"`
	Addresses          []UserAddressResponse  `json:"address,omitempty"`
}

type UserProviderResponse struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Provider       domain.Provider `json:"provider"`
	ProviderUserID *string         `json:"provider_user_id,omitempty"`
}

type UserAddressResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Label        string  `json:"label"`
	Recipient    string  `json:"receipent"`
	Phone        string  `json:"phone"`
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 *string `json:"address_line_2,omitempty"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	PostalCode   string  `json:"postal_code"`
	Country      string  `json:"country"`
	IsPrimary    bool    `json:"is_primary"`
}

type TokenResponse struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type LoginUserResponse struct {
	Token TokenResponse `json:"token"`
	User  UserResponse  `json:"user"`
}