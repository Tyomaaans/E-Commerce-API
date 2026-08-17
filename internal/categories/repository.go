package categories

import (
	"context"

	"gorm.io/gorm"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/pkg"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

// ==== Category Repository ====

func (r *categoryRepository) CreateCategory(ctx context.Context, category *domain.CategoryEntity) (*domain.CategoryEntity, error) {
	storage := ToCategoryStorage(category)

	if err := r.db.WithContext(ctx).Create(storage).Error; err != nil {
		return nil, pkg.HandleDBError(err)
	}
	return ToCategoryEntity(storage), nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, update *domain.UpdateCategoryEntity) error {
	fields := make(map[string]any)

	if update.Name != nil && *update.Name != "" {
		fields["name"] = *update.Name
	}
	if update.Slug != nil && *update.Slug != "" {
		fields["slug"] = *update.Slug
	}
	if update.Image != nil && *update.Image != "" {
		fields["image"] = *update.Slug
	}
	if update.ParentID != nil && *update.ParentID != "" {
		fields["parent_id"] = *update.Slug
	}

	result := r.db.WithContext(ctx).
		Model(&postgres.CategoryStorage{}).
		Where("id = ?", update.ID).
		Updates(fields)

	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *categoryRepository) GetCategories(ctx context.Context) ([]*domain.CategoryEntity, error) {
	var storages []postgres.CategoryStorage

	err := r.db.WithContext(ctx).
		Order("created_at ASC").
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToCategoryListEntity(storages), nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*domain.CategoryEntity, error) {
	var storage postgres.CategoryStorage

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToCategoryEntity(&storage), nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id string) error {
	result := r.db.Unscoped().WithContext(ctx).
		Where("id = ?", id).
		Delete(&postgres.CategoryStorage{})
	
	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}