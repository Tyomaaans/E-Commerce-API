package users

import (
	"context"

	"gorm.io/gorm"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/pkg"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// ==== User Repository ====

func (r *userRepository) CreateUser(ctx context.Context, user *domain.CreateUserEntity) (*domain.UserEntity, error) {
	storage := &postgres.UserStorage{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}

	if err := r.db.WithContext(ctx).Create(storage).Error; err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserEntity(storage), nil
}

// UpdateUser applies a partial update — only non-nil, non-empty fields are written.
func (r *userRepository) UpdateUser(ctx context.Context, update *domain.UpdateUserEntity) error {
	fields := make(map[string]any)

	if update.Name != nil && *update.Name != "" {
		fields["name"] = *update.Name
	}
	if update.Email != nil && *update.Email != "" {
		fields["email"] = *update.Email
	}
	if update.Username != nil && *update.Username != "" {
		fields["username"] = *update.Username
	}
	if update.Phone != nil && *update.Phone != "" {
		fields["phone"] = *update.Phone
	}
	if update.Role != nil && *update.Role != "" {
		fields["role"] = *update.Role
	}
	if update.Avatar != nil && *update.Avatar != "" {
		fields["avatar"] = *update.Avatar
	}
	if update.EmailVerifiedAt != nil {
		fields["email_verified_at"] = update.EmailVerifiedAt
	}
	if update.ProfileCompletedAt != nil {
		fields["profile_completed_at"] = update.ProfileCompletedAt
	}
	if update.RememberMe != nil {
		fields["remember_me"] = *update.RememberMe
	}

	if len(fields) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&postgres.UserStorage{}).
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

func (r *userRepository) UpdatePassword(ctx context.Context, update *domain.UpdatePasswordEntity) error {
	result := r.db.WithContext(ctx).
		Model(&postgres.UserStorage{}).
		Where("email = ?", update.Email).
		Update("password", update.Password)

	if result.Error != nil {
		return pkg.HandleDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pkg.HandleDBError(gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *userRepository) GetUsers(ctx context.Context) ([]domain.UserEntity, error) {
	var storages []postgres.UserStorage

	if err := r.db.WithContext(ctx).
		Preload("Providers").
		Preload("Addresses", "is_primary = ?", true).
		Order("created_at ASC").
		Find(&storages).Error; err != nil {
		return nil, pkg.HandleDBError(err)
	}

	users := make([]domain.UserEntity, 0, len(storages))
	for i := range storages {
		users = append(users, *ToUserEntity(&storages[i]))
	}

	return users, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.UserEntity, error) {
	var storage postgres.UserStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Preload("Providers").
		Preload("Addresses", "is_primary = ?", true).
		Where("id = ?", id).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserEntity(&storage), nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserEntity, error) {
	var storage postgres.UserStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Preload("Providers").
		Preload("Addresses", "is_primary = ?", true).
		Where("email = ?", email).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserEntity(&storage), nil
}

// GetPasswordByEmail fetches only the columns needed for authentication,
// avoiding loading the full user row on every login attempt.
func (r *userRepository) GetPasswordByEmail(ctx context.Context, email string) (string, string, error) {
	var storage postgres.UserStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Select("id", "password").
		Where("email = ?", email).
		First(&storage).Error

	if err != nil {
		return "", "", pkg.HandleDBError(err)
	}

	return storage.Password, storage.ID, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.UserEntity, error) {
	var storage postgres.UserStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Preload("Providers").
		Preload("Addresses").
		Preload("Store").
		Where("username = ?", username).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserEntity(&storage), nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).Delete(&postgres.UserStorage{})
		if result.Error != nil {
			return pkg.HandleDBError(result.Error)
		}
		if result.RowsAffected == 0 {
			return pkg.HandleDBError(gorm.ErrRecordNotFound)
		}

		return nil
	})
}

// ==== User Provider (OAuth) Repository ====

func (r *userRepository) CreateUserProvider(ctx context.Context, provider *domain.UserProviderEntity) error {
	storage := ToUserProviderStorage(provider)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(storage).Error; err != nil {
			return pkg.HandleDBError(err)
		}

		return nil
	})
}

func (r *userRepository) GetUserProvidersByUserID(ctx context.Context, userID string) ([]domain.UserProviderEntity, error) {
	var storages []postgres.UserProviderStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Where("user_id = ?", userID).
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	providers := make([]domain.UserProviderEntity, 0, len(storages))
	for i := range storages {
		providers = append(providers, *ToUserProviderEntity(&storages[i]))
	}

	return providers, nil
}

// GetUserByProvider resolves a user via their OAuth identity (e.g. google, github).
// Joins through user_providers so we don't need a separate lookup step.
func (r *userRepository) GetUserByProvider(ctx context.Context, provider, providerUserID string) (*domain.UserEntity, error) {
	var storage postgres.UserStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Joins("JOIN user_providers ON user_providers.user_id = users.id").
		Where("user_providers.provider = ? AND user_providers.provider_user_id = ?", provider, providerUserID).
		Preload("Providers").
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserEntity(&storage), nil
}

func (r *userRepository) DeleteUserProvider(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).Delete(&postgres.UserProviderStorage{})
		if result.Error != nil {
			return pkg.HandleDBError(result.Error)
		}
		if result.RowsAffected == 0 {
			return pkg.HandleDBError(gorm.ErrRecordNotFound)
		}

		return nil
	})
}

// ==== User Address Repository ====

// CreateAddress persists a new address. If marked as primary, any existing
// primary for the same user is demoted first within the same transaction.
func (r *userRepository) CreateAddress(ctx context.Context, address *domain.UserAddressEntity) (*domain.UserAddressEntity, error) {
	storage := ToUserAddressStorage(address)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if storage.IsPrimary == nil || *storage.IsPrimary {
			if err := tx.Model(&postgres.UserAddressStorage{}).
				Where("user_id = ? AND is_primary = ?", storage.UserID, true).
				Update("is_primary", false).Error; err != nil {
				return pkg.HandleDBError(err)
			}
		}

		return pkg.HandleDBError(tx.Create(storage).Error)
	})

	if err != nil {
		return nil, err
	}

	return ToUserAddressEntity(storage), nil
}

// UpdateAddress applies a partial update to an address. Promoting to primary
// demotes the previous one atomically to prevent duplicate primary addresses.
func (r *userRepository) UpdateAddress(ctx context.Context, update *domain.UpdateUserAddressEntity) error {
	fields := make(map[string]any)

	if update.Label != nil && *update.Label != "" {
		fields["label"] = *update.Label
	}
	if update.Recipient != nil && *update.Recipient != "" {
		fields["recipient"] = *update.Recipient
	}
	if update.AddressLine1 != nil && *update.AddressLine1 != "" {
		fields["address_line1"] = *update.AddressLine1
	}
	if update.AddressLine2 != nil {
		fields["address_line2"] = update.AddressLine2
	}
	if update.City != nil && *update.City != "" {
		fields["city"] = *update.City
	}
	if update.State != nil && *update.State != "" {
		fields["state"] = *update.State
	}
	if update.PostalCode != nil && *update.PostalCode != "" {
		fields["postal_code"] = *update.PostalCode
	}
	if update.Country != nil && *update.Country != "" {
		fields["country"] = *update.Country
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if update.IsPrimary != nil && *update.IsPrimary {
			fields["is_primary"] = true

			if err := tx.Model(&postgres.UserAddressStorage{}).
				Where("user_id = ? AND id != ? AND is_primary = ?", update.UserID, update.ID, true).
				Update("is_primary", false).Error; err != nil {
				return pkg.HandleDBError(err)
			}
		}

		if len(fields) == 0 {
			return nil
		}

		result := tx.Model(&postgres.UserAddressStorage{}).
			Where("id = ? AND user_id = ?", update.ID, update.UserID).
			Updates(fields)

		if result.Error != nil {
			return pkg.HandleDBError(result.Error)
		}
		if result.RowsAffected == 0 {
			return pkg.HandleDBError(gorm.ErrRecordNotFound)
		}

		return nil
	})
}

func (r *userRepository) GetAddresses(ctx context.Context) ([]domain.UserAddressEntity, error) {
	var storages []postgres.UserAddressStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Order("created_at ASC").
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	addresses := make([]domain.UserAddressEntity, 0, len(storages))
	for i := range storages {
		addresses = append(addresses, *ToUserAddressEntity(&storages[i]))
	}

	return addresses, nil
}

func (r *userRepository) GetAddressByID(ctx context.Context, id string) (*domain.UserAddressEntity, error) {
	var storage postgres.UserAddressStorage

	err := r.db.WithContext(ctx).
		Scopes(ActiveUser()).
		Where("id = ?", id).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToUserAddressEntity(&storage), nil
}

// GetAddressesByUserID returns all addresses for a user, primary first then newest.
func (r *userRepository) GetAddressesByUserID(ctx context.Context, id string) ([]domain.UserAddressEntity, error) {
	var storages []postgres.UserAddressStorage

	err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Order("is_primary DESC, created_at DESC").
		Find(&storages).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	entities := make([]domain.UserAddressEntity, 0, len(storages))
	for i := range storages {
		entities = append(entities, *ToUserAddressEntity(&storages[i]))
	}

	return entities, nil
}

// DeleteAddress soft-deletes an address, scoped to userID to prevent cross-tenant deletion.
func (r *userRepository) DeleteAddress(ctx context.Context, addressID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var address postgres.UserAddressStorage
		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).
			First(&address).Error; err != nil {
			return pkg.HandleDBError(err)
		}

		if *address.IsPrimary {
			var count int64
			if err := tx.Model(&postgres.UserAddressStorage{}).
				Where("user_id = ? AND id != ?", userID, addressID).
				Count(&count).Error; err != nil {
				return pkg.HandleDBError(err)
			}

			if count > 0 {
				return pkg.ErrDeletePrimaryAddress
			}
		}

		result := tx.Unscoped().Where("id = ? AND user_id = ?", addressID, userID).
			Delete(&postgres.UserAddressStorage{})

		if result.Error != nil {
			return pkg.HandleDBError(result.Error)
		}
		if result.RowsAffected == 0 {
			return pkg.HandleDBError(gorm.ErrRecordNotFound)
		}

		return nil
	})
}

// ActiveUser is a reusable GORM scope that filters out soft-deleted records.
func ActiveUser() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at IS NULL")
	}
}