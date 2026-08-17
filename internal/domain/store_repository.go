package domain

import (
	"context"
)

type StoreRepository interface {
	CreateStore(ctx context.Context, store *StoreEntity) (*StoreEntity, error)
	UpdateStore(ctx context.Context, update *UpdateStoreEntity) error
	GetStores(ctx context.Context) ([]*StoreEntity, error)
	GetStoreByID(ctx context.Context, id string) (*StoreEntity, error)
	GetStoreByUserID(ctx context.Context, id string) (*StoreEntity, error)
	DeactivateStore(ctx context.Context, id string) error
	ReactivateStore(ctx context.Context, id string) error
	GetInactiveStores(ctx context.Context) ([]*StoreEntity, error)
	DeleteStore(ctx context.Context, id string) error
}