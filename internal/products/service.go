package products

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/pkg"

	jsonValidator "E-COMMERCE-API/pkg"
)

type ProductService interface {
	AddProduct(ctx context.Context, userID string, req AddProductRequest) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, userID string, req *UpdateProductRequest, isAdmin bool) (*ProductResponse, error)
	GetProducts(ctx context.Context) ([]ProductResponse, error)
	GetProductByID(ctx context.Context, id string) (*ProductResponse, error)
	DeactivateProduct(ctx context.Context, userID, productID string, isAdmin bool) error
	ReactivateProduct(ctx context.Context, userID, productID string, isAdmin bool) error
	DeleteProduct(ctx context.Context, id string) error
}

type productService struct {
	productRepo domain.ProductRepository
	storeRepo   domain.StoreRepository
	validate    *validator.Validate
}

func NewProductService(
	productRepo domain.ProductRepository,
	storeRepo domain.StoreRepository,
	validate *validator.Validate,
) ProductService {
	return &productService{
		productRepo: productRepo,
		storeRepo:   storeRepo,
		validate:    validate,
	}
}

func (s *productService) AddProduct(ctx context.Context, userID string, req AddProductRequest) (*ProductResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	slug := strings.ToLower(req.Name)
	req.Slug = strings.ReplaceAll(slug, " ", "-")

	store, err := s.storeRepo.GetStoreByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.ErrForbidden
		}
		return nil, err
	}

	req.StoreID = store.ID

	id := uuid.NewString()
	payload := ToAddProductEntity(id, &req)

	addProduct, err := s.productRepo.CreateProduct(ctx, payload)
	if err != nil {
		return nil, err
	}

	res := ToProductResponse(addProduct)
	if err := formatResponseIDs(res, addProduct.ID, addProduct.StoreID); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *productService) UpdateProduct(ctx context.Context, userID string, req *UpdateProductRequest, isAdmin bool) (*ProductResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	decodedProductID, err := base64ToUUID(req.ID)
	if err != nil {
		return nil, err
	}
	req.ID = decodedProductID

	// Merchant must own the store associated with the product
	if !isAdmin {
		store, err := s.storeRepo.GetStoreByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.ErrForbidden
			}
			return nil, err
		}
		req.StoreID = store.ID
	}

	if req.Name != nil {
		slug := strings.ToLower(*req.Name)
		slug = strings.ReplaceAll(slug, " ", "-")
		req.Slug = &slug
	}

	update := ToUpdateProductEntity(req)
	if err := s.productRepo.UpdateProduct(ctx, update); err != nil {
		return nil, err
	}

	updated, err := s.productRepo.GetProductByID(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	res := ToProductResponse(updated)
	if err := formatResponseIDs(res, updated.ID, updated.StoreID); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *productService) GetProducts(ctx context.Context) ([]ProductResponse, error) {
	products, err := s.productRepo.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, pkg.ErrNotFound
	}

	res := ToProductListResponse(products)
	for i := range products {
		productID, err := uuidToBase64(products[i].ID)
		if err != nil {
			return nil, err
		}

		storeID, err := uuidToBase64(products[i].StoreID)
		if err != nil {
			return nil, err
		}

		res[i].ID = productID
		res[i].StoreID = storeID
	}

	return res, nil
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*ProductResponse, error) {
	decodedID, err := base64ToUUID(id)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetProductByID(ctx, decodedID)
	if err != nil {
		return nil, err
	}

	res := ToProductResponse(product)
	if err := formatResponseIDs(res, product.ID, product.StoreID); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *productService) DeactivateProduct(ctx context.Context, userID, productID string, isAdmin bool) error {
	decodedProductID, err := base64ToUUID(productID)
	if err != nil {
		return err
	}

	if !isAdmin {
		if err := s.validateProductOwnership(ctx, userID, decodedProductID); err != nil {
			return err
		}
	}

	return s.productRepo.DeactivateProduct(ctx, decodedProductID)
}

func (s *productService) ReactivateProduct(ctx context.Context, userID, productID string, isAdmin bool) error {
	decodedProductID, err := base64ToUUID(productID)
	if err != nil {
		return err
	}

	if !isAdmin {
		if err := s.validateProductOwnership(ctx, userID, decodedProductID); err != nil {
			return err
		}
	}

	return s.productRepo.ReactivateProduct(ctx, decodedProductID)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	decodedID, err := base64ToUUID(id)
	if err != nil {
		return err
	}

	return s.productRepo.DeleteProduct(ctx, decodedID)
}

// Internal helpers

func (s *productService) validateProductOwnership(ctx context.Context, userID, decodedProductID string) error {
	store, err := s.storeRepo.GetStoreByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrForbidden
		}
		return err
	}

	product, err := s.productRepo.GetProductByID(ctx, decodedProductID)
	if err != nil {
		return err
	}

	if product.StoreID != store.ID {
		return pkg.ErrForbidden
	}

	return nil
}

func formatResponseIDs(res *ProductResponse, rawProductID, rawStoreID string) error {
	productID, err := uuidToBase64(rawProductID)
	if err != nil {
		return err
	}

	storeID, err := uuidToBase64(rawStoreID)
	if err != nil {
		return err
	}

	res.ID = productID
	res.StoreID = storeID
	return nil
}

func uuidToBase64(uuidStr string) (string, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(parsed[:]), nil
}

func base64ToUUID(encoded string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	parsed, err := uuid.FromBytes(decoded)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}