package categories

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"E-COMMERCE-API/internal/domain"

	jsonValidator "E-COMMERCE-API/pkg"
)

// Service Category Interface ====

type CategoryService interface {
	AddParentCategory(ctx context.Context, req *AddCategoryParentRequest) (*CategoryResponse, error)
	AddChildCategory(ctx context.Context, req *AddCategoryChildRequest) (*CategoryResponse, error)
	UpdateParentCategory(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error)
	ChangeParentCategory(ctx context.Context, req *MoveCategoryParentRequest) (*CategoryResponse, error)
	GetCategories(ctx context.Context) ([]*CategoryResponse, error)
	GetCategoryByID(ctx context.Context, id string) (*CategoryResponse, error)
	DeleteCategory(ctx context.Context, id string) error
}

type categoryService struct {
	categoryRepo domain.CategoryRepository
	validate     *validator.Validate
}

func NewCategoryService (
	categoryRepo domain.CategoryRepository,
	validate     *validator.Validate,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		validate:     validate,
	}
}

func (s *categoryService) AddParentCategory(ctx context.Context, req *AddCategoryParentRequest) (*CategoryResponse, error) {
	slug := strings.ToLower(req.Name)
	req.Slug = strings.ReplaceAll(slug, " ", "-")

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	ID      := uuid.NewString()
	payload := AddParentCategoryToEntity(ID, req)

	addCategoryParent, err := s.categoryRepo.CreateCategory(ctx, payload)
	if err != nil {
		return nil, err
	}

	return ToCategoryResponse(addCategoryParent), nil
}

func (s *categoryService) AddChildCategory(ctx context.Context, req *AddCategoryChildRequest) (*CategoryResponse, error) {
	slug := strings.ToLower(req.Name)
	req.Slug = strings.ReplaceAll(slug, " ", "-")

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	parent, err := s.categoryRepo.GetCategoryByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	req.ParentID = parent.ID

	ID      := uuid.NewString()
	payload := AddChildCategoryToEntity(ID, req)

	addCategoryChild, err := s.categoryRepo.CreateCategory(ctx, payload)
	if err != nil {
		return nil, err
	}

	return ToCategoryResponse(addCategoryChild), nil
}

func (s *categoryService) UpdateParentCategory(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	if req.Name != nil {
		slug := strings.ToLower(*req.Name)
		slug  = strings.ReplaceAll(slug, " ", "-")
		req.Slug = &slug
	}

	update := UpdateCategoryRequestToEntity(req)
	if err := s.categoryRepo.UpdateCategory(ctx, update); err != nil {
		return nil, err
	}

	updated, err := s.categoryRepo.GetCategoryByID(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	return ToCategoryResponse(updated), nil
}

func (s *categoryService) ChangeParentCategory(ctx context.Context, req *MoveCategoryParentRequest) (*CategoryResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	update := UpdateCategoryRequestToEntity(&UpdateCategoryRequest{
		ID:       req.ID,
		ParentID: req.ParentID,
	})

	if err := s.categoryRepo.UpdateCategory(ctx, update); err != nil {
		return nil, err
	}

	updated, err := s.categoryRepo.GetCategoryByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return ToCategoryResponse(updated), nil
}

func (s *categoryService) GetCategories(ctx context.Context) ([]*CategoryResponse, error) {
	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	tree := buildCategoryTree(categories)

	return ToCategoryListResponse(tree), nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*CategoryResponse, error) {
	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ToCategoryResponse(category), nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	if err := s.categoryRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	
	return nil
}

// ==== Internal Helper ====

func buildCategoryTree(categories []*domain.CategoryEntity) []*domain.CategoryEntity {
    mapped := make(map[string]*domain.CategoryEntity, len(categories))
    for _, cat := range categories {
        mapped[cat.ID] = cat
    }

    var roots []*domain.CategoryEntity

    for _, cat := range categories {
        if cat.ParentID == nil {
            roots = append(roots, cat)
        } else {
            if parent, ok := mapped[*cat.ParentID]; ok {
                parent.Children = append(parent.Children, cat)
            }
        }
    }

    return roots
}