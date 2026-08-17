package stores

import (
	"context"

	"gorm.io/gorm"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/pkg"
)

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) domain.StoreRepository {
	return &storeRepository{db: db}
}

// ==== Store Repository ====

func (r *storeRepository) CreateStore(ctx context.Context, store *domain.StoreEntity) (*domain.StoreEntity, error) {
	storage := ToStoreStorage(store)

	if err := r.db.WithContext(ctx).Create(storage).Error; err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToStoreEntity(storage), nil
}

func (r *storeRepository) UpdateStore(ctx context.Context, update *domain.UpdateStoreEntity) error {
	fields := make(map[string]any)

	if update.Name != nil && *update.Name != "" {
		fields["name"] = *update.Name
	}
	if update.Slug != nil && *update.Slug != "" {
		fields["slug"] = *update.Slug
	}
	if update.Description != nil && *update.Description != "" {
		fields["description"] = *update.Description
	}
	if update.Logo != nil && *update.Logo != "" {
		fields["logo"] = *update.Logo
	}

	if len(fields) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&postgres.StoreStorage{}).
		Where("id = ? AND user_id = ?", update.ID, update.UserID).
		Updates(fields)

	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *storeRepository) GetStores(ctx context.Context) ([]*domain.StoreEntity, error) {
	var storages []postgres.StoreStorage

	err := r.db.WithContext(ctx).
		Order("created_at ASC").
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToStoreListEntity(storages), nil
}

func (r *storeRepository) GetStoreByID(ctx context.Context, id string) (*domain.StoreEntity, error) {
	var storage postgres.StoreStorage

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToStoreEntity(&storage), nil
}

func (r *storeRepository) GetStoreByUserID(ctx context.Context, id string) (*domain.StoreEntity, error) {
	var storage postgres.StoreStorage

	err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Preload("User").
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToStoreEntity(&storage), nil
}

func (r *storeRepository) DeactivateStore(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        result := tx.Model(&postgres.StoreStorage{}).
            Where("id = ?", id).
            Update("is_active", false)

        if result.Error != nil {
            return pkg.HandleDBError(result.Error)
        }
        if result.RowsAffected == 0 {
            return pkg.HandleDBError(gorm.ErrRecordNotFound)
        }

        return tx.Delete(&postgres.StoreStorage{}, "id = ?", id).Error
    })
}

func (r *storeRepository) ReactivateStore(ctx context.Context, id string) error {
	result := r.db.Unscoped().WithContext(ctx).
		Model(&postgres.StoreStorage{}).
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

func (r *storeRepository) GetInactiveStores(ctx context.Context) ([]*domain.StoreEntity, error) {
	var storages []postgres.StoreStorage

	err := r.db.WithContext(ctx).
		Unscoped().Where("deleted_at IS NOT NULL").
		Order("created_at ASC").
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToStoreListEntity(storages), nil
}

func (r *storeRepository) DeleteStore(ctx context.Context, id string) error {
	result := r.db.Unscoped().WithContext(ctx).
		Where("id = ?", id).
		Delete(&postgres.StoreStorage{})
	
	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}