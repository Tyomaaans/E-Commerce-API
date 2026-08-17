package domain

import (
	"context"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *ProductEntity) (*ProductEntity, error)
	UpdateProduct(ctx context.Context, update *UpdateProductEntity) error
	GetProducts(ctx context.Context) ([]ProductEntity, error)
	GetProductByID(ctx context.Context, id string) (*ProductEntity, error)
	GetProductBySlug(ctx context.Context, slug string) (*ProductEntity, error)
	GetProductByIDAndSlug(ctx context.Context, id, slug string) (*ProductEntity, error)
	DeactivateProduct(ctx context.Context, id string) error
	ReactivateProduct(ctx context.Context, id string) error
	DeleteProduct(ctx context.Context, id string) error
}