package domain

import (
	"context"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *CategoryEntity) (*CategoryEntity, error)
	UpdateCategory(ctx context.Context, update *UpdateCategoryEntity) error
	GetCategories(ctx context.Context) ([]*CategoryEntity, error)
	GetCategoryByID(ctx context.Context, id string) (*CategoryEntity, error)
	DeleteCategory(ctx context.Context, id string) error
}