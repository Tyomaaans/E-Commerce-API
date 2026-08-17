package stores

import (
	"errors"
	"net/http"
	
	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/pkg"
)

type StoreHandler struct {
	storeService StoreService
}

func NewStoreHandler(storeService StoreService) *StoreHandler {
	return &StoreHandler{
		storeService: storeService,
	}
}

func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkg.ErrForbidden):
		pkg.ErrorResponse(c, http.StatusForbidden, err)
	case errors.Is(err, pkg.ErrNotFound):
		pkg.ErrorResponse(c, http.StatusNotFound, errors.New("store not found or inactive"))
	case errors.Is(err, pkg.ErrInternal):
		pkg.ErrorResponse(c, http.StatusInternalServerError, err)
	default:
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
	}
}

func (h *StoreHandler) RegisterStore(c *gin.Context) {
	var req RegisterStoreRequest

	req.UserID = c.GetString("userID")

	val, exists := c.Get("userCache")
	if exists {
		userCache := val.(*domain.UserCache)
		if userCache.Role == domain.Seller {
			pkg.ErrorResponse(c, http.StatusForbidden, nil)
			return
		}
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	store, err := h.storeService.RegisterStore(c.Request.Context(), req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "registered store successfully", map[string]interface{}{
		"store": store,
	})
}

func (h *StoreHandler) UpdateStore(c *gin.Context) {
	var req UpdateStoreRequest

	req.UserID = c.GetString("userID")

	val, exists := c.Get("userCache")
	if exists {
		userCache := val.(*domain.UserCache)
		if userCache.Role == domain.Buyer {
			pkg.ErrorResponse(c, http.StatusForbidden, nil)
			return
		}
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	store, err := h.storeService.UpdateStore(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated store successfully", map[string]interface{}{
		"store": store,
	})

}

func (h *StoreHandler) UpdateUserStore(c *gin.Context) {
	var req UpdateStoreRequest

	req.UserID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	store, err := h.storeService.UpdateStore(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated store successfully", map[string]interface{}{
		"store": store,
	})

}

func (h *StoreHandler) GetStores(c *gin.Context) {
	stores, err := h.storeService.GetStores(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get stores successfully", map[string]interface{}{
		"store": stores,
	})
}

func (h *StoreHandler) GetStoreByID(c *gin.Context) {
	id := c.Param("id")

	store, err := h.storeService.GetStoreByID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get store successfully", map[string]interface{}{
		"store": store,
	})
}

func (h *StoreHandler) GetStoreByUserID(c *gin.Context) {
	id := c.GetString("userID")

	store, err := h.storeService.GetStoreByUserID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get store successfully", map[string]interface{}{
		"store": store,
	})
}

func (h *StoreHandler) GetUserStoreByUserID(c *gin.Context) {
	id := c.Param("id")

	store, err := h.storeService.GetStoreByUserID(c.Request.Context(), id)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get user store successfully", map[string]interface{}{
		"store": store,
	})
}

func (h *StoreHandler) DeactivateStore(c *gin.Context) {
	id := c.GetString("userID")

	if err := h.storeService.DeactivateStore(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deactivated store successfully", nil)
}

func (h *StoreHandler) DeactivateUserStore(c *gin.Context) {
	id := c.Param("id")

	if err := h.storeService.DeactivateStore(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deactivated user store successfully", nil)
}

func (h *StoreHandler) ReactivateUserStore(c *gin.Context) {
	id := c.Param("id")

	if err := h.storeService.ReactivateStore(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "reactivated user store successfully", nil)
}

func (h *StoreHandler) DeleteUserStore(c *gin.Context) {
	id := c.Param("id")

	if err := h.storeService.DeleteStore(c.Request.Context(), id); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted user store successfully", nil)
}

func (h *StoreHandler) GetInactiveStores(c *gin.Context) {
	stores, err := h.storeService.GetInactiveStores(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get inactive stores successfully", map[string]interface{}{
		"store": stores,
	})
}