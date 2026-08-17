package categories

import (
	"errors"
	"net/http"
	
	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/pkg"
)

type CategoryHandler struct {
	categoryService CategoryService
}

func NewCategoryHandler(categoryService CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkg.ErrNotFound):
		pkg.ErrorResponse(c, http.StatusNotFound, errors.New("category not found"))
	case errors.Is(err, pkg.ErrInternal):
		pkg.ErrorResponse(c, http.StatusInternalServerError, err)
	default:
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
	}
}

func (h *CategoryHandler) AddParentCategory(c *gin.Context) {
	var req AddCategoryParentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	
	category, err := h.categoryService.AddParentCategory(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "added parent category successfully", map[string]interface{}{
		"category": category,
	})
}

func (h *CategoryHandler) AddChildCategory(c *gin.Context) {
	var req AddCategoryChildRequest

	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	
	category, err := h.categoryService.AddChildCategory(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "added child category successfully", map[string]interface{}{
		"category": category,
	})
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var req UpdateCategoryRequest

	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	
	updated, err := h.categoryService.UpdateParentCategory(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated category successfully", map[string]interface{}{
		"category": updated,
	})
}

func (h *CategoryHandler) ChangeParentCategory(c *gin.Context) {
	var req MoveCategoryParentRequest

	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	
	updated, err := h.categoryService.ChangeParentCategory(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated category successfully", map[string]interface{}{
		"category": updated,
	})
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetCategories(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get categories successfully", map[string]interface{}{
		"categories": categories,
	})
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get category successfully", map[string]interface{}{
		"category": category,
	})
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	if err := h.categoryService.DeleteCategory(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted category successfully", nil)
}