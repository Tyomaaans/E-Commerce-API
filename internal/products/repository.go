package products

import (
	"context"

	"gorm.io/gorm"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/pkg"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

// ==== Product Repository ====

func (r *productRepository) CreateProduct(ctx context.Context, product *domain.ProductEntity) (*domain.ProductEntity, error) {
	storage := ToProductStorage(product)

	if err := r.db.WithContext(ctx).Create(storage).Error; err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToProductEntity(storage), nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, update *domain.UpdateProductEntity) error {
	fields := make(map[string]any)

	if update.CategoryID != nil && *update.CategoryID != "" {
		fields["category_id"] = *update.CategoryID
	}
	if update.Name != nil && *update.Name != "" {
		fields["name"] = *update.Name
	}
	if update.Slug != nil && *update.Slug != "" {
		fields["slug"] = *update.Slug
	}
	if update.Description != nil && *update.Description != "" {
		fields["description"] = *update.Description
	}
	if update.Price != nil {
		fields["price"] = *update.Price
	}
	if update.Stock != nil {
		fields["stock"] = *update.Stock
	}
	if update.WeightGram != nil {
		fields["weight_gram"] = *update.WeightGram
	}

	if len(fields) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&postgres.ProductStorage{}).
		Where("id = ? AND store_id = ?", update.ID, update.StoreID).
		Updates(fields)

	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *productRepository) GetProducts(ctx context.Context) ([]domain.ProductEntity, error) {
	var products []postgres.ProductStorage

	err := r.db.WithContext(ctx).Unscoped().
		Preload("Images").
		Preload("Reviews").
		Order("created_at ASC").
		Find(&products).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToProductListEntity(products), nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*domain.ProductEntity, error) {
	var product postgres.ProductStorage

	err := r.db.WithContext(ctx).Unscoped().
		Where("id = ?", id).
		Preload("Images").
		Preload("Reviews").
		First(&product).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToProductEntity(&product), nil
}

func (r *productRepository) GetProductBySlug(ctx context.Context, slug string) (*domain.ProductEntity, error) {
	var product postgres.ProductStorage

	err := r.db.WithContext(ctx).Unscoped().
		Where("slug = ?", slug).
		Preload("Images").
		Preload("Reviews").
		First(&product).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToProductEntity(&product), nil
}

func (r *productRepository) GetProductByIDAndSlug(ctx context.Context, id, slug string) (*domain.ProductEntity, error) {
	var product postgres.ProductStorage

	err := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND slug = ?", id, slug).
		Preload("Images").
		Preload("Reviews").
		First(&product).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToProductEntity(&product), nil
}

func (r *productRepository) DeactivateProduct(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        result := tx.Model(&postgres.ProductStorage{}).
            Where("id = ?", id).
            Update("is_active", false)

        if result.Error != nil {
            return pkg.HandleDBError(result.Error)
        }
        if result.RowsAffected == 0 {
            return pkg.HandleDBError(gorm.ErrRecordNotFound)
        }

        return tx.Delete(&postgres.ProductStorage{}, "id = ?", id).Error
    })
}

func (r *productRepository) ReactivateProduct(ctx context.Context, id string) error {
	result := r.db.Unscoped().WithContext(ctx).
		Model(&postgres.ProductStorage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
            "is_active":  true, 
            "deleted_at": nil,
        })

	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	result := r.db.Unscoped().WithContext(ctx).
		Where("id = ?", id).
		Delete(&postgres.ProductStorage{})
	
	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}