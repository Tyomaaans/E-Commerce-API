package stores

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/users"
	"E-COMMERCE-API/pkg"

	jsonValidator "E-COMMERCE-API/pkg"
)

// ==== Service Store Interface ====

type StoreService interface{
	RegisterStore(ctx context.Context, req RegisterStoreRequest) (*StoreResponse, error)
	UpdateStore(ctx context.Context, req *UpdateStoreRequest) (*StoreResponse, error)
	GetStores(ctx context.Context) ([]*StoreResponse, error)
	GetStoreByID(ctx context.Context, id string) (*StoreResponse, error)
	GetStoreByUserID(ctx context.Context, id string) (*StoreResponse, error)
	DeactivateStore(ctx context.Context, id string) error
	ReactivateStore(ctx context.Context, id string) error
	GetInactiveStores(ctx context.Context) ([]*StoreResponse, error)
	DeleteStore(ctx context.Context, id string) error
}

type storeService struct {
	storeRepo domain.StoreRepository
	userRepo  domain.UserRepository
	cacheRepo domain.UserCacheRepository
	validate  *validator.Validate
}

func NewStoreService (
	storeRepo domain.StoreRepository,
	userRepo  domain.UserRepository,
	cacheRepo domain.UserCacheRepository,
	validate  *validator.Validate,
) StoreService {
	return &storeService{
		storeRepo: storeRepo,
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		validate:  validate,
	}
}

func (s *storeService) RegisterStore(ctx context.Context, req RegisterStoreRequest) (*StoreResponse, error) {
	slug := strings.ToLower(req.Name)
	req.Slug = strings.ReplaceAll(slug, " ", "-")

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	ID      := uuid.NewString()
	payload := StoreRequestToEntity(ID, &req)

	registerStore, err := s.storeRepo.CreateStore(ctx, payload)
	if err != nil {
		return nil, err
	}
	 
	seller := domain.Seller
	if err := s.userRepo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:   req.UserID,
		Role: &seller,
	}); err != nil {
		_ = s.storeRepo.DeactivateStore(ctx, registerStore.ID)
		_ = s.storeRepo.DeleteStore(ctx, registerStore.ID)
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	_ = s.cacheRepo.SetUserCache(ctx, users.ToUserCache(user))

	return ToStoreResponse(registerStore), nil
}

func (s *storeService) UpdateStore(ctx context.Context, req *UpdateStoreRequest) (*StoreResponse, error) {
	store, err := s.storeRepo.GetStoreByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.ErrForbidden
		}
		return nil, err
	}

	req.ID = store.ID

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	if req.Name != nil {
		slug := strings.ToLower(*req.Name)
		slug  = strings.ReplaceAll(slug, " ", "-")
		req.Slug = &slug
	}

	if err := s.storeRepo.UpdateStore(ctx, UpdateStoreRequestToEntity(req)); err != nil {
		return nil, err
	}

	updated, err := s.storeRepo.GetStoreByID(ctx, store.ID)
	if err != nil {
		return nil, err
	}

	return ToStoreResponse(updated), nil
}

func (s *storeService) GetStores(ctx context.Context) ([]*StoreResponse, error) {
	stores, err := s.storeRepo.GetStores(ctx)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, err
		}
		return nil, err
	}

	if len(stores) == 0 {
		return nil, pkg.ErrNotFound
	}

	for i := range stores {
		stores[i].UserID = ""
	}

	return ToStoreListResponse(stores), nil
}

func (s *storeService) GetStoreByID(ctx context.Context, id string) (*StoreResponse, error) {
	store, err := s.storeRepo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}

	store.UserID = ""

	return ToStoreResponse(store), nil
}

func (s *storeService) GetStoreByUserID(ctx context.Context, id string) (*StoreResponse, error) {
	store, err := s.storeRepo.GetStoreByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.ErrForbidden
		}
	}

	return ToStoreResponse(store), nil
}

func (s *storeService) DeactivateStore(ctx context.Context, id string) error {
	store, err := s.storeRepo.GetStoreByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrForbidden
		}
	}

	if err := s.storeRepo.DeactivateStore(ctx, store.ID); err != nil {
		return err
	}

	buyer := domain.Buyer
	if err := s.userRepo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:   id,
		Role: &buyer,
	}); err != nil {
		_ = s.storeRepo.ReactivateStore(ctx, store.ID)
		return err
	}

	return nil
}

func (s *storeService) ReactivateStore(ctx context.Context, id string) error {
	stores, err := s.storeRepo.GetInactiveStores(ctx)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrNotFound
		}
	}

	var storeID string
	for i := range stores {
		if stores[i].UserID == id {
			storeID = stores[i].ID
		}
	}

	if err := s.storeRepo.ReactivateStore(ctx, storeID); err != nil {
		return err
	}

	seller := domain.Seller
	if err := s.userRepo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:   id,
		Role: &seller,
	}); err != nil {
		_ = s.storeRepo.DeactivateStore(ctx, storeID)
		return err
	}

	return nil
}

func (s *storeService) GetInactiveStores(ctx context.Context) ([]*StoreResponse, error) {
	stores, err := s.storeRepo.GetInactiveStores(ctx)
	if err != nil {
		return nil, err
	}

	return ToStoreListResponse(stores), nil
}

func (s *storeService) DeleteStore(ctx context.Context, id string) error {
	store, err := s.storeRepo.GetStoreByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrNotFound
		}
	}

	if err := s.storeRepo.DeleteStore(ctx, store.ID); err != nil {
		return err
	}

	buyer := domain.Buyer
	if err := s.userRepo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:   id,
		Role: &buyer,
	}); err != nil {
		return err
	}

	return nil
}