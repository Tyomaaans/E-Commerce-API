package stores

import (
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/internal/products"
	"E-COMMERCE-API/internal/users"
)

// ==== Store To Storage <-> Entity ====

func ToStoreEntity(s *postgres.StoreStorage) *domain.StoreEntity {
	if s == nil {
		return nil
	}

	store := &domain.StoreEntity{
		ID:          s.ID,
		UserID:      s.UserID,
		Name:        s.Name,
		Slug:        s.Slug,
		Description: s.Description,
		Logo:        s.Logo,
		IsActive:    s.IsActive,
	}

	if s.User != nil {
		store.User = users.ToUserEntity(s.User)
	}

	if len(s.Products) > 0 {
		store.Products = products.ToProductListEntity(s.Products)
	}

	return store
}

func ToStoreStorage(e *domain.StoreEntity) *postgres.StoreStorage {
	if e == nil {
		return nil
	}

	return &postgres.StoreStorage{
		ID:          e.ID,
		UserID:      e.UserID,
		Name:        e.Name,
		Slug:        e.Slug,
		Description: e.Description,
		Logo:        e.Logo,
		IsActive:    e.IsActive,
	}
}

func ToStoreListEntity(list []postgres.StoreStorage) []*domain.StoreEntity {
	result := make([]*domain.StoreEntity, len(list))

	for i := range list {
		result[i] = ToStoreEntity(&list[i])
	}

	return result
}

// ==== Request To Entity ====

func StoreRequestToEntity(id string, req *RegisterStoreRequest) *domain.StoreEntity {
	if req == nil {
		return nil
	}

	return &domain.StoreEntity{
		ID:          id,
		UserID:      req.UserID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Logo:        req.Logo,
		IsActive:    req.IsActive,
	}
}

func UpdateStoreRequestToEntity(req *UpdateStoreRequest) *domain.UpdateStoreEntity {
	if req == nil {
		return nil
	}

	return &domain.UpdateStoreEntity{
		ID:          req.ID,
		UserID:      req.UserID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Logo:        req.Logo,
	}
}

// ==== Entity To Response ====

func ToStoreResponse(e *domain.StoreEntity) *StoreResponse {
	if e == nil {
		return nil
	}

	res := &StoreResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		Name:        e.Name,
		Slug:        e.Slug,
		Description: e.Description,
		Logo:        e.Logo,
		IsActive:    e.IsActive,
	}

	if e.User != nil {
		res.User = users.ToUserResponse(e.User)
	}

	if len(e.Products) > 0 {
		res.Products = products.ToProductListResponse(e.Products)
	}

	return res
}

func ToStoreListResponse(list []*domain.StoreEntity) []*StoreResponse {
	result := make([]*StoreResponse, len(list))

	for i, e := range list {
		result[i] = ToStoreResponse(e)
	}

	return result
}