package domain

import (
	"context"
)

type UserRepository interface {
    // User
    CreateUser(ctx context.Context, user *CreateUserEntity) (*UserEntity, error)
    UpdateUser(ctx context.Context, update *UpdateUserEntity) error
    UpdatePassword(ctx context.Context, update *UpdatePasswordEntity) error
    GetUsers(ctx context.Context) ([]UserEntity, error)
    GetUserByID(ctx context.Context, id string) (*UserEntity, error)
    GetUserByEmail(ctx context.Context, email string) (*UserEntity, error)
    GetPasswordByEmail(ctx context.Context, email string) (string, string, error)
    GetUserByUsername(ctx context.Context, username string) (*UserEntity, error)
    DeleteUser(ctx context.Context, id string) error

    // Provider
    CreateUserProvider(ctx context.Context, provider *UserProviderEntity) error
    GetUserProvidersByUserID(ctx context.Context, id string) ([]UserProviderEntity, error)
    GetUserByProvider(ctx context.Context, provider, providerUserID string) (*UserEntity, error)
    DeleteUserProvider(ctx context.Context, id string) error

    // Address
    CreateAddress(ctx context.Context, address *UserAddressEntity) (*UserAddressEntity, error)
    UpdateAddress(ctx context.Context, update *UpdateUserAddressEntity) error
    GetAddresses(ctx context.Context) ([]UserAddressEntity, error)
    GetAddressByID(ctx context.Context, id string) (*UserAddressEntity, error)
    GetAddressesByUserID(ctx context.Context, id string) ([]UserAddressEntity, error)
    DeleteAddress(ctx context.Context, addressID, userID string) error
}