package products

import (
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
)

// ==== Product To Storage <-> Entity ====

func ToProductEntity(s *postgres.ProductStorage) *domain.ProductEntity {
	product := &domain.ProductEntity{
		ID:            s.ID,
		StoreID:       s.StoreID,
		CategoryID:    s.CategoryID,
		Name:          s.Name,
		Slug:          s.Slug,
		Description:   s.Description,
		Price:         s.Price,
		Stock:         s.Stock,
		ReservedStock: s.ReservedStock,
		WeightGram:    s.WeightGram,
		IsActive:      s.IsActive,
	}

	if len(s.Images) > 0 {
		product.Images = ToProductImageListEntity(s.Images)
	}

	if len(s.Reviews) > 0 {
		product.Reviews = ToProductReviewListEntity(s.Reviews)
	}

	return product
}

func ToProductStorage(e *domain.ProductEntity) *postgres.ProductStorage {
	return &postgres.ProductStorage{
		ID:            e.ID,
		StoreID:       e.StoreID,
		CategoryID:    e.CategoryID,
		Name:          e.Name,
		Slug:          e.Slug,
		Description:   e.Description,
		Price:         e.Price,
		Stock:         e.Stock,
		ReservedStock: e.ReservedStock,
		WeightGram:    e.WeightGram,
		IsActive:      e.IsActive,
	}
}

func ToProductListEntity(list []postgres.ProductStorage) []domain.ProductEntity {
	result := make([]domain.ProductEntity, len(list))

	for i := range list {
		result[i] = *ToProductEntity(&list[i])
	}

	return result
}

// ==== Product Image To Storage <-> Entity ====

func ToProductImageEntity(s *postgres.ProductImageStorage) *domain.ProductImageEntity {
	return &domain.ProductImageEntity{
		ID:        s.ID,
		ProductID: s.ProductID,
		ImageURL:  s.ImageURL,
		IsPrimary: s.IsPrimary,
	}
}

func ToProductImageStorage(e *domain.ProductImageEntity) *postgres.ProductImageStorage {
	return &postgres.ProductImageStorage{
		ID:        e.ID,
		ProductID: e.ProductID,
		ImageURL:  e.ImageURL,
		IsPrimary: e.IsPrimary,
	}
}

func ToProductImageListEntity(list []postgres.ProductImageStorage) []domain.ProductImageEntity {
	result := make([]domain.ProductImageEntity, len(list))

	for i := range list {
		result[i] = *ToProductImageEntity(&list[i])
	}

	return result
}

// ==== Product Review To Storage <-> Entity ====

func ToProductReviewEntity(s *postgres.ProductReviewStorage) *domain.ProductReviewEntity {
	if s == nil {
		return nil
	}

	return &domain.ProductReviewEntity{
		ID:        s.ID,
		UserID:    s.UserID,
		ProductID: s.ProductID,
		Rating:    s.Rating,
		Comment:   s.Comment,
	}
}

func ToProductReviewStorage(e *domain.ProductReviewEntity) *postgres.ProductReviewStorage {
	if e == nil {
		return nil
	}

	return &postgres.ProductReviewStorage{
		ID:        e.ID,
		UserID:    e.UserID,
		ProductID: e.ProductID,
		Rating:    e.Rating,
		Comment:   e.Comment,
	}
}

func ToProductReviewListEntity(list []postgres.ProductReviewStorage) []domain.ProductReviewEntity {
	result := make([]domain.ProductReviewEntity, len(list))

	for i := range list {
		result[i] = *ToProductReviewEntity(&list[i])
	}

	return result
}

// ==== Request To Entity ====

func ToAddProductEntity(id string, req *AddProductRequest) *domain.ProductEntity {
	return &domain.ProductEntity{
		ID:          id,
		StoreID:     req.StoreID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		WeightGram:  req.WeightGram,
		IsActive:    true,
	}
}

func ToUpdateProductEntity(req *UpdateProductRequest) *domain.UpdateProductEntity {
	return &domain.UpdateProductEntity{
		ID:          req.ID,
		StoreID:     req.StoreID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		WeightGram:  req.WeightGram,
	}
}

func ToAddProductImageEntity(id, productID string, req *AddProductImageRequest) *domain.ProductImageEntity {
	return &domain.ProductImageEntity{
		ID:        id,
		ProductID: productID,
		ImageURL:  req.ImageURL,
		IsPrimary: req.IsPrimary,
	}
}

func ToUpdateProductImageEntity(id, productID string, req *UpdateProductImageRequest) *domain.UpdateProductImageEntity {
	return &domain.UpdateProductImageEntity{
		ID:        id,
		ProductID: productID,
		ImageURL:  req.ImageURL,
		IsPrimary: req.IsPrimary,
	}
}

func ToAddProductReviewEntity(id, productID string, req *AddProductReviewRequest) *domain.ProductReviewEntity {
	return &domain.ProductReviewEntity{
		ID:        id,
		UserID:    req.UserID,
		ProductID: productID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
}

func ToUpdateProductReviewEntity(productID string, req *UpdateProductReviewRequest) *domain.UpdateReviewEntity {
	return &domain.UpdateReviewEntity{
		ID:        req.ID,
		UserID:    req.UserID,
		ProductID: productID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
}

// ==== Entity To Response ====

func ToProductResponse(e *domain.ProductEntity) *ProductResponse {
	response := &ProductResponse{
		ID:            e.ID,
		StoreID:       e.StoreID,
		CategoryID:    e.CategoryID,
		Name:          e.Name,
		Slug:          e.Slug,
		Description:   e.Description,
		Price:         e.Price,
		Stock:         e.Stock,
		ReservedStock: e.ReservedStock,
		WeightGram:    e.WeightGram,
		IsActive:      e.IsActive,
	}

	if len(e.Images) > 0 {
		response.Images = ToProductImageListResponse(e.Images)
	}

	if len(e.Reviews) > 0 {
		response.Reviews = ToProductReviewListResponse(e.Reviews)
	}

	return response
}

func ToProductListResponse(list []domain.ProductEntity) []ProductResponse {
	result := make([]ProductResponse, len(list))

	for i := range list {
		result[i] = *ToProductResponse(&list[i])
	}

	return result
}

func ToProductImageResponse(e *domain.ProductImageEntity) *ProductImageResponse {
	return &ProductImageResponse{
		ID:        e.ID,
		ProductID: e.ProductID,
		ImageURL:  e.ImageURL,
		IsPrimary: e.IsPrimary,
	}
}

func ToProductImageListResponse(list []domain.ProductImageEntity) []ProductImageResponse {
	result := make([]ProductImageResponse, len(list))

	for i := range list {
		result[i] = *ToProductImageResponse(&list[i])
	}

	return result
}

func ToProductReviewResponse(e *domain.ProductReviewEntity) *ProductReviewResponse {
	return &ProductReviewResponse{
		ID:        e.ID,
		UserID:    e.UserID,
		ProductID: e.ProductID,
		Rating:    e.Rating,
		Comment:   e.Comment,
	}
}

func ToProductReviewListResponse(list []domain.ProductReviewEntity) []ProductReviewResponse {
	result := make([]ProductReviewResponse, len(list))

	for i := range list {
		result[i] = *ToProductReviewResponse(&list[i])
	}

	return result
}