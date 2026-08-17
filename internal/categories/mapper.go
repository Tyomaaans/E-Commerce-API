package categories

import (
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/infrastructure/postgres"
)

// ==== Category To Storage <-> Entity ====

func ToCategoryEntity(s *postgres.CategoryStorage) *domain.CategoryEntity {
	if s == nil {
		return nil
	}

	category := &domain.CategoryEntity{
		ID:       s.ID,
		Name:     s.Name,
		Slug:     s.Slug,
		Image:    s.Image,
		ParentID: s.ParentID,
	}

	return category
}

func ToCategoryStorage(e *domain.CategoryEntity) *postgres.CategoryStorage {
	if e == nil {
		return nil
	}

	return &postgres.CategoryStorage{
		ID:       e.ID,
		Name:     e.Name,
		Slug:     e.Slug,
		Image:    e.Image,
		ParentID: e.ParentID,
	}
}

func ToCategoryListEntity(list []postgres.CategoryStorage) []*domain.CategoryEntity {
	result := make([]*domain.CategoryEntity, len(list))

	for i := range list {
		result[i] = ToCategoryEntity(&list[i])
	}

	return result
}

// ==== Request To Entity ====

func AddParentCategoryToEntity(id string, req *AddCategoryParentRequest) *domain.CategoryEntity {
	if req == nil {
		return nil
	}

	return &domain.CategoryEntity{
		ID:       id,
		Name:     req.Name,
		Slug:     req.Slug,
		Image:    req.Image,
	}
}

func AddChildCategoryToEntity(id string, req *AddCategoryChildRequest) *domain.CategoryEntity {
	if req == nil {
		return nil
	}

	return &domain.CategoryEntity{
		ID:       id,
		Name:     req.Name,
		Slug:     req.Slug,
		Image:    req.Image,
		ParentID: &req.ParentID,
	}
}

func UpdateCategoryRequestToEntity(req *UpdateCategoryRequest) *domain.UpdateCategoryEntity {
	if req == nil {
		return nil
	}

	return &domain.UpdateCategoryEntity{
		ID:       req.ID,
		Name:     req.Name,
		Slug:     req.Slug,
		Image:    req.Image,
		ParentID: &req.ParentID,
	}
}

// ==== Entity To Ressponse ====

func ToCategoryResponse(e *domain.CategoryEntity) *CategoryResponse {
	if e == nil {
		return nil
	}

	res := &CategoryResponse{
		ID:       e.ID,
		Name:     e.Name,
		Slug:     e.Slug,
		Image:    e.Image,
		ParentID: e.ParentID,
	}

	for _, child := range e.Children {
        res.Children = append(res.Children, ToCategoryResponse(child))
    }

	return res
}

func ToCategoryListResponse(list []*domain.CategoryEntity) []*CategoryResponse {
	result := make([]*CategoryResponse, 0, len(list))

	for _, e := range list {
		result = append(result, ToCategoryResponse(e))
	}

	return result
}