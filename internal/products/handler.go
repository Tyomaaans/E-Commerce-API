package products

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/pkg"
)

type ProductHandler struct {
	productService ProductService
}

func NewProductHandler(productService ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkg.ErrForbidden):
		pkg.ErrorResponse(c, http.StatusForbidden, errors.New("store not found or access denied"))
	case errors.Is(err, pkg.ErrNotFound):
		pkg.ErrorResponse(c, http.StatusNotFound, errors.New("product not found or inactive"))
	case errors.Is(err, pkg.ErrInternal):
		pkg.ErrorResponse(c, http.StatusInternalServerError, err)
	default:
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
	}
}

func (h *ProductHandler) AddProduct(c *gin.Context) {
	var req AddProductRequest
	userID := c.GetString("userID")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	product, err := h.productService.AddProduct(c.Request.Context(), userID, req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "added product successfully", map[string]interface{}{
		"product": product,
	})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	var req UpdateProductRequest

	userID := c.GetString("userID")
	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), userID, &req, false)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated product successfully", map[string]interface{}{
		"product": product,
	})
}

func (h *ProductHandler) AdminUpdateProduct(c *gin.Context) {
	var req UpdateProductRequest

	userID := c.Param("userID")
	req.ID  = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), userID, &req, true)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated product successfully", map[string]interface{}{
		"product": product,
	})
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	products, err := h.productService.GetProducts(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get products successfully", map[string]interface{}{
		"products": products,
	})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	product, err := h.productService.GetProductByID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get product successfully", map[string]interface{}{
		"product": product,
	})
}

func (h *ProductHandler) DeactivateProduct(c *gin.Context) {
	userID    := c.GetString("userID")
	productID := c.Param("id")

	if err := h.productService.DeactivateProduct(c.Request.Context(), userID, productID, false); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deactivated product successfully", nil)
}

func (h *ProductHandler) AdminDeactivateProduct(c *gin.Context) {
	userID    := c.Param("userID")
	productID := c.Param("id")

	if err := h.productService.DeactivateProduct(c.Request.Context(), userID, productID, true); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deactivated product successfully", nil)
}

func (h *ProductHandler) ReactivateProduct(c *gin.Context) {
	userID    := c.GetString("userID")
	productID := c.Param("id")

	if err := h.productService.ReactivateProduct(c.Request.Context(), userID, productID, false); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "reactivated product successfully", nil)
}

func (h *ProductHandler) AdminReactivateProduct(c *gin.Context) {
	userID    := c.Param("userID")
	productID := c.Param("id")

	if err := h.productService.ReactivateProduct(c.Request.Context(), userID, productID, true); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "reactivated product successfully", nil)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted product successfully", nil)
}