package domain

import (
	"time"
)

type Role string

const (
    Seller Role = "seller"
    Buyer  Role = "buyer"
    Admin  Role = "admin"
)

type UserEntity struct {
    ID                 string
    Name               *string
    Email              string
    Username           *string
    Phone              *string
    Role               Role
    Avatar             *string
    EmailVerifiedAt    *time.Time
    ProfileCompletedAt *time.Time
    RememberMe         bool

    Providers []UserProviderEntity
    Addresses []UserAddressEntity
    Store     *StoreEntity
}

type CreateUserEntity struct {
    ID       string
    Email    string
    Password string
    Role     Role
}

type UpdateUserEntity struct {
    ID                 string
    Name               *string
    Email              *string
    Username           *string
    Phone              *string
    Role               *Role
    Avatar             *string
    EmailVerifiedAt    *time.Time
    ProfileCompletedAt *time.Time
    RememberMe         *bool
}

type UpdatePasswordEntity struct {
    Email    string
    Password string
}