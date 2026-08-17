package users

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/pkg"
)

type UserHandler struct {
	userService          UserService
	defaultRefreshExpiry time.Duration
	shortRefreshExpiry   time.Duration
}

func NewUserHandler(userService UserService, defaultRefreshExpiry time.Duration, shortRefreshExpiry time.Duration) *UserHandler {
	return &UserHandler{
		userService:          userService,
		defaultRefreshExpiry: defaultRefreshExpiry,
		shortRefreshExpiry:   shortRefreshExpiry,
	}
}

// httpError maps domain-level errors to the appropriate HTTP status code and
// writes the error response. Centralising this logic avoids scattering status
// code decisions across every handler.
func httpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pkg.ErrNotFound):
		pkg.ErrorResponse(c, http.StatusNotFound, err)
	case errors.Is(err, pkg.ErrInvalidCredentials),
		errors.Is(err, pkg.ErrInvalidPassword):
		pkg.ErrorResponse(c, http.StatusUnauthorized, err)
	case errors.Is(err, pkg.ErrInternal):
		pkg.ErrorResponse(c, http.StatusInternalServerError, err)
	default:
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
	}
}

// setRefreshTokenCookie writes the refresh token into an HTTP-only cookie.
// The MaxAge is driven by the user's rememberMe preference — longer for
// persistent sessions, shorter for ephemeral ones.
func (h *UserHandler) setRefreshTokenCookie(c *gin.Context, token string, rememberMe bool) {
	expiryDuration := h.defaultRefreshExpiry
	if !rememberMe {
		expiryDuration = h.shortRefreshExpiry
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		token,
		int(expiryDuration/time.Second),
		"/api/v1/auth",
		"",
		false,
		true,
	)
}

// clearRefreshTokenCookie expires the refresh token cookie on the client side
// by setting MaxAge to -1, effectively forcing immediate deletion.
func (h *UserHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/v1/auth",
		"",
		false,
		true,
	)
}

// ==== User Auth Handlers ====

// Register handles new user sign-up. If the email exists but is unverified,
// the service re-issues a verification email and a 200 is returned instead of 201.
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.userService.Register(c.Request.Context(), req); err != nil {
		// ErrEmailSentAgain is not a failure — surface it as a 200 with a helpful message.
		if errors.Is(err, pkg.ErrEmailSentAgain) {
			pkg.SuccessResponse(
				c,
				http.StatusOK,
				"email was already registered but not verified. a new verification link has been sent to your email.",
				nil,
			)
			return
		}

		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "register user successfully. please check your email for verification", nil)
}

// ResendVerification re-sends the email verification link for unverified accounts.
// Returns 404 if the email is not registered and 400 if it is already verified.
func (h *UserHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.userService.ResendVerificationEmail(c.Request.Context(), req); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.ErrorResponse(c, http.StatusNotFound, errors.New("account not registered yet, please register first to continue."))
			return
		}
		if errors.Is(err, pkg.ErrEmailAlreadyVerified) {
			pkg.ErrorResponse(c, http.StatusBadRequest, pkg.ErrEmailAlreadyVerified)
			return
		}

		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "verification email has been resent successfully", nil)
}

// VerifyRegister validates the email token from the query string, marks the account
// as verified, and returns an access token along with a refresh token cookie.
func (h *UserHandler) VerifyRegister(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		pkg.ErrorResponse(c, http.StatusBadRequest, pkg.ErrInvalidOrExpiredToken)
		return
	}

	res, err := h.userService.VerifyRegister(c.Request.Context(), token)
	if err != nil {
		// An invalid or expired token is an auth failure, not a bad request.
		httpError(c, err)
		return
	}

	// New accounts are not on "remember me" by default.
	h.setRefreshTokenCookie(c, res.Token.RefreshToken, false)

	pkg.SuccessResponse(c, http.StatusOK, "verified register successfully", map[string]interface{}{
		"user":  res.User,
		"token": res.Token.AccessToken,
	})
}

// Login authenticates a user and issues a JWT pair. The refresh token is set as an
// HTTP-only cookie while the access token is returned in the response body.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	res, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		// Wrong credentials and unverified email are both auth-level failures.
		httpError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, res.Token.RefreshToken, res.User.RememberMe)

	pkg.SuccessResponse(c, http.StatusOK, "logged in successfully", map[string]interface{}{
		"user":  res.User,
		"token": res.Token.AccessToken,
	})
}

// RefreshToken reads the refresh token from the cookie, rotates the token pair,
// and sets the new refresh token back into the cookie with the correct TTL.
func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Missing cookie means the user is not authenticated.
		pkg.ErrorResponse(c, http.StatusUnauthorized, err)
		return
	}

	res, isRememberMe, err := h.userService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		httpError(c, err)
		return
	}

	// Preserve the original rememberMe TTL when rotating the cookie.
	h.setRefreshTokenCookie(c, res.RefreshToken, isRememberMe)

	pkg.SuccessResponse(c, http.StatusOK, "refresh token successfully", map[string]interface{}{
		"token": res.AccessToken,
	})
}

// Logout revokes both tokens and clears the refresh token cookie so the client
// is fully signed out across all surfaces that rely on the cookie.
func (h *UserHandler) Logout(c *gin.Context) {
	accessToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// No cookie means there is nothing to revoke — treat as unauthenticated.
		pkg.ErrorResponse(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.userService.Logout(c.Request.Context(), accessToken, refreshToken); err != nil {
		httpError(c, err)
		return
	}

	h.clearRefreshTokenCookie(c)

	pkg.SuccessResponse(c, http.StatusOK, "logged out successfully", nil)
}

// ForgotPassword accepts an email address and triggers the reset flow.
// The response is intentionally vague to avoid leaking whether an account exists.
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.userService.ForgotPassword(c.Request.Context(), req); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "if email is registered, password reset instructions have been sent", nil)
}

// VerifyResetPassword checks whether a reset token is still valid.
// Used by the frontend to enable the "set new password" step before the user
// types anything — avoids a failed submission after form fill.
func (h *UserHandler) VerifyResetPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		pkg.ErrorResponse(c, http.StatusBadRequest, pkg.ErrInvalidOrExpiredToken)
		return
	}

	if err := h.userService.VerifyResetPassword(c.Request.Context(), token); err != nil {
		// An invalid token is a 401 — the client should redirect to forgot-password.
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "token is valid", nil)
}

// ResetPassword consumes a valid reset token and replaces the user's password.
// The token is burned server-side after a successful update.
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), req); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "reset password successfully", nil)
}

// ==== User Management Handlers ====

// GetUsers returns a list of all users. Restricted to admin roles via middleware.
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetUsers(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get users successfully", map[string]interface{}{
		"users": users,
	})
}

// GetMyProfile retrieves the authenticated user's own profile using the userID
// injected by the auth middleware.
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get profile successfully", map[string]interface{}{
		"user": user,
	})
}

// GetUserByID fetches any user by their UUID path param. Restricted to admin roles.
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get user by id successfully", map[string]interface{}{
		"user": user,
	})
}

// UpdateUser allows an admin to update any user's fields. The target user ID is
// bound from the URI path and merged with the request body before being passed to the service.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest

	if err := c.ShouldBindUri(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated user successfully", map[string]interface{}{
		"user": updatedUser,
	})
}

// UpdateMyProfile lets an authenticated user update their own non-sensitive profile fields.
// The user ID is bound from the URI and must match the authenticated session.
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	var req UpdateMySelfRequest

	if err := c.ShouldBindUri(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	updatedUser, err := h.userService.UpdateMyProfile(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "updated user successfully", map[string]interface{}{
		"user": updatedUser,
	})
}

// CompleteProfile handles the post-registration onboarding step. If the cache already
// shows the profile as completed, the request is short-circuited with a 200.
func (h *UserHandler) CompleteProfile(c *gin.Context) {
	var req CompleteProfileRequest

	// Populate the user ID and guard against re-completion using the middleware cache.
	val, exists := c.Get("userCache")
	if exists {
		userCache := val.(*domain.UserCache)
		req.UserID = userCache.ID
		if userCache.IsProfileCompleted {
			pkg.SuccessResponse(c, http.StatusOK, "profile is already completed", nil)
			return
		}
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	updatedUser, err := h.userService.CompleteProfile(c.Request.Context(), req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "completed profile successfully", map[string]interface{}{
		"user": updatedUser,
	})
}

// ChangePassword lets an authenticated user update their password after verifying
// the current one. The email is pulled from the middleware cache to avoid spoofing.
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePassword

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Override whatever email the client sends — always use the one from the verified session.
	val, exists := c.Get("userCache")
	if exists {
		userCache := val.(*domain.UserCache)
		req.Email = userCache.Email
	}

	if err := h.userService.UpdatePassword(c.Request.Context(), req); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "changed password successfully", nil)
}

// DeleteUser permanently removes a user by ID. Admin-only action.
// Associated cache and session data are evicted as part of the service call.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted user successfully", nil)
}

// DeleteAccount lets a user permanently remove their own account. Both tokens are
// revoked immediately and the refresh token cookie is left to expire naturally —
// the next request will find it invalid regardless.
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID      := c.GetString("userID")
	accessToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// Cannot revoke tokens without the refresh token — reject early.
		pkg.ErrorResponse(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), userID, accessToken, refreshToken); err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted account successfully", nil)
}

// ==== User Address Handlers ====

func (h *UserHandler) AddAddress(c *gin.Context) {
	var req AddAddressRequest

	userID := c.GetString("userID")

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	req.UserID = userID

	address, err := h.userService.AddAddress(c.Request.Context(), req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusCreated, "added address successfully", map[string]interface{}{
		"address": address,
	})
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	var req UpdateAddressRequest

	userID := c.GetString("userID")

	if err := c.ShouldBindUri(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	req.UserID = userID

	address, err := h.userService.UpdateAddress(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "update address successfully", map[string]interface{}{
		"address": address,
	})
}

func (h *UserHandler) UpdateUserAddress(c *gin.Context) {
	var req UpdateAddressRequest

	userID := c.Param("userID")

	if err := c.ShouldBindUri(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	req.UserID = userID

	address, err := h.userService.UpdateAddress(c.Request.Context(), &req)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "update address successfully", map[string]interface{}{
		"address": address,
	})
}

func (h *UserHandler) GetAddresses(c *gin.Context) {
	addresses, err := h.userService.GetAddresses(c.Request.Context())
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get addresses successfully", map[string]interface{}{
		"address": addresses,
	})
}

func (h *UserHandler) GetAddressByID(c *gin.Context) {
	addressID := c.Param("id")
	
	address, err := h.userService.GetAddressByID(c.Request.Context(), addressID)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get address successfully", map[string]interface{}{
		"address": address,
	})
}

func (h *UserHandler) GetAddressesByUserID(c *gin.Context) {
	userID := c.GetString("userID")
	
	addresses, err := h.userService.GetAddressesByUserID(c.Request.Context(), userID)
	if err != nil {
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "get addresses successfully", map[string]interface{}{
		"addresses": addresses,
	})
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	addressID := c.Param("id")
	userID    := c.GetString("userID")

	if err := h.userService.DeleteAddress(c.Request.Context(), addressID, userID); err != nil{
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted address successfully", nil)
}

func (h *UserHandler) DeleteUserAddress(c *gin.Context) {
	userID    := c.Param("userID")
	addressID := c.Param("id")

	if err := h.userService.DeleteAddress(c.Request.Context(), addressID, userID); err != nil{
		httpError(c, err)
		return
	}

	pkg.SuccessResponse(c, http.StatusOK, "deleted user address successfully", nil)
}