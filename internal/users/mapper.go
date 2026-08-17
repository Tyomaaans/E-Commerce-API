package users

import (
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/internal/jwt"
	"time"
)

// ==== User To Storage <-> Entity ====

func ToUserEntity(s *postgres.UserStorage) *domain.UserEntity {
	if s == nil {
		return nil
	}

	u := &domain.UserEntity{
		ID:                 s.ID,
		Name:               s.Name,
		Email:              s.Email,
		Username:           s.Username,
		Phone:              s.Phone,
		Role:               s.Role,
		Avatar:             s.Avatar,
		EmailVerifiedAt:    s.EmailVerifiedAt,
		ProfileCompletedAt: s.ProfileCompletedAt,
		RememberMe:         s.RememberMe,
	}

	if len(s.Providers) > 0 {
		u.Providers = ToUserProviderListEntity(s.Providers)
	}

	if len(s.Addresses) > 0 {
		u.Addresses = ToUserAddressListEntity(s.Addresses)
	}

	return u
}

func ToUserStorage(e *domain.UserEntity) *postgres.UserStorage {
	if e == nil {
		return nil
	}

	return &postgres.UserStorage{
		ID:                 e.ID,
		Name:               e.Name,
		Email:              e.Email,
		Username:           e.Username,
		Phone:              e.Phone,
		Role:               e.Role,
		Avatar:             e.Avatar,
		EmailVerifiedAt:    e.EmailVerifiedAt,
		ProfileCompletedAt: e.ProfileCompletedAt,
		RememberMe:         e.RememberMe,
	}
}

func ToUserListEntity(list []postgres.UserStorage) []*domain.UserEntity {
	result := make([]*domain.UserEntity, len(list))

	for i := range list {
		result[i] = ToUserEntity(&list[i])
	}

	return result
}

// ==== User Provider To Storage <-> Entity ====

func ToUserProviderEntity(s *postgres.UserProviderStorage) *domain.UserProviderEntity {
	if s == nil {
		return nil
	}

	return &domain.UserProviderEntity{
		ID:             s.ID,
		UserID:         s.UserID,
		Provider:       s.Provider,
		ProviderUserID: s.ProviderUserID,
	}
}

func ToUserProviderStorage(e *domain.UserProviderEntity) *postgres.UserProviderStorage {
	if e == nil {
		return nil
	}

	return &postgres.UserProviderStorage{
		ID:             e.ID,
		UserID:         e.UserID,
		Provider:       e.Provider,
		ProviderUserID: e.ProviderUserID,
	}
}

func ToUserProviderListEntity(list []postgres.UserProviderStorage) []domain.UserProviderEntity {
	result := make([]domain.UserProviderEntity, len(list))

	for i := range list {
		entity := ToUserProviderEntity(&list[i])
		if entity != nil {
			result[i] = *entity
		}
	}

	return result
}

// ==== User Address To Storage <-> Entity ====

func ToUserAddressEntity(s *postgres.UserAddressStorage) *domain.UserAddressEntity {
	if s == nil {
		return nil
	}

	u := &domain.UserAddressEntity{
		ID:           s.ID,
		UserID:       s.UserID,
		Label:        s.Label,
		Recipient:    s.Recipient,
		Phone:        s.Phone,
		AddressLine1: s.AddressLine1,
		AddressLine2: s.AddressLine2,
		City:         s.City,
		State:        s.State,
		PostalCode:   s.PostalCode,
		Country:      s.Country,
		IsPrimary:    s.IsPrimary,
	}

	return u
}

func ToUserAddressStorage(e *domain.UserAddressEntity) *postgres.UserAddressStorage {
	if e == nil {
		return nil
	}

	return &postgres.UserAddressStorage{
		ID:           e.ID,
		UserID:       e.UserID,
		Label:        e.Label,
		Recipient:    e.Recipient,
		Phone:        e.Phone,
		AddressLine1: e.AddressLine1,
		AddressLine2: e.AddressLine2,
		City:         e.City,
		State:        e.State,
		PostalCode:   e.PostalCode,
		Country:      e.Country,
		IsPrimary:    e.IsPrimary,
	}
}

func ToUserAddressListEntity(list []postgres.UserAddressStorage) []domain.UserAddressEntity {
	result := make([]domain.UserAddressEntity, len(list))

	for i := range list {
		entity := ToUserAddressEntity(&list[i])
		if entity != nil {
			result[i] = *entity
		}
	}

	return result
}

// ==== Cache To Entity ====

func ToUserCache(user *domain.UserEntity) *domain.UserCache {
	providers := make([]string, len(user.Providers))
	for i, p := range user.Providers {
		providers[i] = string(p.Provider)
	}

	return &domain.UserCache{
		ID:                 user.ID,
		Email:              user.Email,
		Role:               user.Role,
		IsEmailVerified:    user.EmailVerifiedAt != nil,
		EmailVerifiedAt:    user.EmailVerifiedAt,
		IsProfileCompleted: user.ProfileCompletedAt != nil,
		ProfileCompletedAt: user.ProfileCompletedAt,
		Providers:          providers,
	}
}

// ==== Request To Entity ====

func ToRegisterUserRequest(id string, req RegisterUserRequest) *domain.CreateUserEntity {
	return &domain.CreateUserEntity{
		ID:       id,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
}

func ToCompleteProfileRequest(req CompleteProfileRequest) *domain.UpdateUserEntity {
	now := time.Now()

	return &domain.UpdateUserEntity{
		ID:                 req.UserID,
		Name:               &req.Name,
		Username:           &req.Username,
		Phone:              &req.Phone,
		Avatar:             req.Avatar,
		ProfileCompletedAt: &now,
	}
}

func ToAddAddressRequest(id string, req *AddAddressRequest) *domain.UserAddressEntity {
	if req == nil {
		return nil
	}

	return &domain.UserAddressEntity{
		ID:           id,
		UserID:       req.UserID,
		Label:        req.Label,
		Recipient:    req.Recipient,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		IsPrimary:    req.IsPrimary,
	}
}

func ToUpdateUserRequest(req *UpdateUserRequest) *domain.UpdateUserEntity {
	if req == nil {
		return nil
	}

	return &domain.UpdateUserEntity{
		ID:                 req.ID,
		Name:               req.Name,
		Email:              req.Email,
		Username:           req.Username,
		Phone:              req.Phone,
		Role:               req.Role,
		Avatar:             req.Avatar,
		EmailVerifiedAt:    req.EmailVerifiedAt,
		ProfileCompletedAt: req.ProfileCompletedAt,
		RememberMe:         req.RememberMe,
	}
}

func ToUpdateAddressRequest(req *UpdateAddressRequest) *domain.UpdateUserAddressEntity {
	if req == nil {
		return nil
	}
	
	return &domain.UpdateUserAddressEntity{
		ID:           req.ID,
		UserID:       req.UserID,
		Label:        req.Label,
		Recipient:    req.Recipient,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		IsPrimary:    req.IsPrimary,
	}
}

func ToCreateUserProviderRequest(req CreateUserProviderRequest) *domain.CreateUserProviderEntity {
	return &domain.CreateUserProviderEntity{
		UserID:         req.UserID,
		Provider:       req.Provider,
		ProviderUserID: req.ProviderUserID,
		Password:       req.Password,
	}
}

func ToResetPasswordRequest(email string, req ResetPasswordRequest) *domain.UpdatePasswordEntity {
    return &domain.UpdatePasswordEntity{
        Email:    email,
        Password: req.Password,
    }
}

// ==== Entity To Response ====

func ToTokenResponse(t *jwt.TokenPair) *TokenResponse {
	if t == nil {
		return nil
	}

	resp := &TokenResponse{
		UserID:       t.UserID,
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}

	return resp
}

func ToUserResponse(e *domain.UserEntity) *UserResponse {
	if e == nil {
		return nil
	}

	resp := &UserResponse{
		ID:         e.ID,
		Name:       e.Name,
		Email:      e.Email,
		Username:   e.Username,
		Phone:      e.Phone,
		Role:       string(e.Role),
		Avatar:     e.Avatar,
		RememberMe: e.RememberMe,
	}

	if e.EmailVerifiedAt != nil {
		formatted := e.EmailVerifiedAt.Format(timeLayout)
		resp.EmailVerifiedAt = &formatted
	}

	if e.ProfileCompletedAt != nil {
		formatted := e.ProfileCompletedAt.Format(timeLayout)
		resp.ProfileCompletedAt = &formatted
	}

	if len(e.Providers) > 0 {
		resp.Providers = ToUserProviderListResponse(e.Providers)
	}

	if len(e.Addresses) > 0 {
		resp.Addresses = ToUserAddressListResponse(e.Addresses)
	}

	return resp
}

func ToUserListResponse(list []domain.UserEntity) []UserResponse {
	result := make([]UserResponse, len(list))

	for i, e := range list {
		result[i] = *ToUserResponse(&e)
	}

	return result
}

func ToUserProviderResponse(e *domain.UserProviderEntity) *UserProviderResponse {
	if e == nil {
		return nil
	}

	return &UserProviderResponse{
		ID:             e.ID,
		UserID:         e.UserID,
		Provider:       e.Provider,
		ProviderUserID: e.ProviderUserID,
	}
}

func ToUserProviderListResponse(list []domain.UserProviderEntity) []UserProviderResponse {
	result := make([]UserProviderResponse, len(list))

	for i, e := range list {
		result[i] = *ToUserProviderResponse(&e)
	}

	return result
}

func ToUserAddressResponse(e *domain.UserAddressEntity) *UserAddressResponse {
	if e == nil {
		return nil
	}

	return &UserAddressResponse{
		ID:           e.ID,
		UserID:       e.UserID,
		Label:        e.Label,
		Recipient:    e.Recipient,
		Phone:        e.Phone,
		AddressLine1: e.AddressLine1,
		AddressLine2: e.AddressLine2,
		City:         e.City,
		State:        e.State,
		PostalCode:   e.PostalCode,
		Country:      e.Country,
		IsPrimary:    *e.IsPrimary,
	}
}

func ToUserAddressListResponse(list []domain.UserAddressEntity) []UserAddressResponse {
	result := make([]UserAddressResponse, len(list))

	for i, e := range list {
		result[i] = *ToUserAddressResponse(&e)
	}

	return result
}

// ==== Helper Formatting ====

const timeLayout = "2006-01-02 15:04:05"