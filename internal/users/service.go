package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"E-COMMERCE-API/internal/jwt"
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/pkg"

	jsonValidator "E-COMMERCE-API/pkg"
)

// ==== Service Interface ====

type UserService interface {
	// Auth
	Register(ctx context.Context, req RegisterUserRequest) error
	ResendVerificationEmail(ctx context.Context, req ResendVerificationRequest) error
	VerifyRegister(ctx context.Context, token string) (*LoginUserResponse, error)
	Login(ctx context.Context, req LoginUserRequest) (*LoginUserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, bool, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	VerifyResetPassword(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error

	// User CRUD
	GetUsers(ctx context.Context) ([]UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UserResponse, error)
	UpdateMyProfile(ctx context.Context, req *UpdateMySelfRequest) (*UserResponse, error)
	CompleteProfile(ctx context.Context, req CompleteProfileRequest) (*UserResponse, error)
	UpdatePassword(ctx context.Context, req ChangePassword) error
	DeleteUser(ctx context.Context, id string) error
	DeleteAccount(ctx context.Context, id, accessToken, refreshToken string) error

	// User Address
	AddAddress(ctx context.Context, req AddAddressRequest) (*UserAddressResponse, error)
	UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*UserAddressResponse, error)
	GetAddresses(ctx context.Context) ([]UserAddressResponse, error)
	GetAddressByID(ctx context.Context, id string) (*UserAddressResponse, error)
	GetAddressesByUserID(ctx context.Context, id string) ([]UserAddressResponse, error)
	DeleteAddress(ctx context.Context, addressID, userID string) error
}

// ==== Service Implementation ====

type userService struct {
	repo      domain.UserRepository
	tokenRepo UserTokenRepository
	jwtRepo   jwt.JWTService
	producer  EmailProducer
	cacheRepo domain.UserCacheRepository
	validate  *validator.Validate
}

func NewUserService (
	repo domain.UserRepository,
	tokenRepo UserTokenRepository,
	jwtRepo jwt.JWTService,
	producer EmailProducer,
	cacheRepo domain.UserCacheRepository,
	validate *validator.Validate,
) UserService {
	return &userService{
		repo:      repo,
		tokenRepo: tokenRepo,
		jwtRepo:   jwtRepo,
		producer:  producer,
		cacheRepo: cacheRepo,
		validate:  validate,
	}
}

// ==== Auth ====

// Register creates a new pending user and dispatches a verification email via the job queue.
// If the email is already registered but unverified, the existing record is reused and a fresh
// verification token is issued — avoiding duplicate accounts.
func (s *userService) Register(ctx context.Context, req RegisterUserRequest) error {
	req.Role  = domain.Buyer
	req.Email = strings.ToLower(req.Email)

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	// Check whether the email is already taken before doing any heavy work.
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pkg.ErrNotFound) {
		return err
	}

	if existing != nil {
		// A verified account already owns this email — reject outright.
		if existing.EmailVerifiedAt != nil {
			return pkg.ErrEmailAlreadyUsed
		}

		// Account exists but is unverified — resend the token instead of creating a duplicate.
		if err := s.sendRegistrationToken(ctx, existing.ID, existing.Email); err != nil {
			return err
		}

		return pkg.ErrEmailSentAgain
	}

	// Hash the password before persisting anything to the database.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return pkg.ErrInternal
	}

	// Build the create payload with a fresh UUID and the hashed password.
	ID := uuid.NewString()
	req.Password = string(hashedPassword)

	createUserPayload := ToRegisterUserRequest(ID, req)

	createdUser, err := s.repo.CreateUser(ctx, createUserPayload)
	if err != nil {
		return err
	}

	// Generate a verification token and push the email job to the queue.
	return s.sendRegistrationToken(ctx, createdUser.ID, createdUser.Email)
}

// ResendVerificationEmail re-issues a verification token for accounts that haven't confirmed
// their email yet. Intended to back frontend "Resend email" CTA buttons.
func (s *userService) ResendVerificationEmail(ctx context.Context, req ResendVerificationRequest) error {
	req.Email = strings.ToLower(req.Email)

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	// Ensure the account actually exists before proceeding.
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrNotFound
		}
		return err
	}

	// Nothing to resend if the email is already confirmed.
	if user.EmailVerifiedAt != nil {
		return pkg.ErrEmailAlreadyVerified
	}

	// Generate a fresh token and dispatch the verification email.
	return s.sendRegistrationToken(ctx, user.ID, user.Email)
}

// VerifyRegister validates the one-time email token, marks the account as verified,
// and returns a JWT pair so the user is immediately logged in after confirmation.
func (s *userService) VerifyRegister(ctx context.Context, token string) (*LoginUserResponse, error) {
	// Resolve the token to a user ID; reject if missing or expired.
	userID, err := s.tokenRepo.GetRegisterToken(ctx, token)
	if err != nil {
		return nil, pkg.ErrInvalidOrExpiredToken
	}

	// Stamp the email_verified_at timestamp to mark the account as active.
	now := time.Now()
	if err := s.repo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:              userID,
		EmailVerifiedAt: &now,
	}); err != nil {
		return nil, err
	}

	// Token is single-use — remove it from Redis immediately.
	_ = s.tokenRepo.DeleteRegisterToken(ctx, token)

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Warm up the cache so subsequent authenticated requests skip the DB.
	_ = s.cacheRepo.SetUserCache(ctx, ToUserCache(user))

	tokenPair, err := s.jwtRepo.GenerateTokenPair(ctx, user.ID, string(user.Role), user.RememberMe)
	if err != nil {
		return nil, pkg.ErrInternal
	}

	// Derive a display name from the email local-part as a fallback when no name is set.
	var parsedName string
	if addr, err := mail.ParseAddress(user.Email); err == nil {
		parts := strings.Split(addr.Address, "@")
		parsedName = parts[0]
	}

	// Fire the welcome email in the background — failure is non-fatal here.
	s.sendEmailJob(ctx, EmailJob{
		Type: EmailJobRegisterSuccess,
		To:   user.Email,
		Name: parsedName,
	})

	return &LoginUserResponse{
		Token: *ToTokenResponse(tokenPair),
		User:  *ToUserResponse(user),
	}, nil
}

// Login authenticates a user by email and password, then returns a JWT pair on success.
func (s *userService) Login(ctx context.Context, req LoginUserRequest) (*LoginUserResponse, error) {
	req.Email = strings.ToLower(req.Email)
	
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	// Use a generic error for both "not found" and "wrong password" to prevent user enumeration.
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.ErrInvalidCredentials
		}
		return nil, err
	}

	// Block login until the email address is confirmed.
	if user.EmailVerifiedAt == nil {
		return nil, pkg.ErrEmailNotVerified
	}

	hashedPassword, _, err := s.repo.GetPasswordByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	if err := s.repo.UpdateUser(ctx, &domain.UpdateUserEntity{
		ID:         user.ID,
		RememberMe: &req.RememberMe,
	}); err != nil {
		return nil, err
	}

	user.RememberMe = req.RememberMe

	// Populate cache on successful login to avoid a DB hit on the first authenticated request.
	_ = s.cacheRepo.SetUserCache(ctx, ToUserCache(user))

	tokenPair, err := s.jwtRepo.GenerateTokenPair(ctx, user.ID, string(user.Role), user.RememberMe)
	if err != nil {
		return nil, pkg.ErrInternal
	}

	return &LoginUserResponse{
		Token: *ToTokenResponse(tokenPair),
		User:  *ToUserResponse(user),
	}, nil
}

// RefreshToken rotates the token pair using the supplied refresh token.
// It also returns the user's rememberMe flag so the caller can set an
// appropriate cookie TTL without an extra DB round-trip.
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, bool, error) {
	pairToken, err := s.jwtRepo.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, false, err
	}

	user, err := s.repo.GetUserByID(ctx, pairToken.UserID)
	if err != nil {
		return nil, false, err
	}

	return pairToken, user.RememberMe, nil
}

// Logout revokes both the access and refresh tokens, effectively invalidating the session.
func (s *userService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	return s.jwtRepo.RevokeTokens(ctx, accessToken, refreshToken)
}

// ForgotPassword sends a password-reset link to the given email address.
// Non-existent emails are silently ignored to prevent user enumeration attacks.
func (s *userService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	req.Email = strings.ToLower(req.Email)

	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	// Return nil for unknown emails — the response is intentionally ambiguous.
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil
		}
		return err
	}

	// Generate a short-lived reset token (15-minute TTL) and store it in Redis.
	token, err := generateSecureToken()
	if err != nil {
		return pkg.ErrInternal
	}

	if err := s.tokenRepo.SetResetPasswordToken(ctx, token, user.Email, 15*time.Minute); err != nil {
		return err
	}

	// Dispatch the reset-password email job to the queue.
	s.sendEmailJob(ctx, EmailJob{
		Type:  EmailJobResetPassword,
		To:    user.Email,
		Token: token,
	})

	return nil
}

// VerifyResetPassword checks whether a reset token is still valid without consuming it.
// Used by the frontend to gate the "enter new password" UI step.
func (s *userService) VerifyResetPassword(ctx context.Context, token string) error {
	_, err := s.tokenRepo.GetResetPasswordToken(ctx, token)
	if err != nil {
		return pkg.ErrInvalidOrExpiredToken
	}

	return nil
}

// ResetPassword validates the reset token, hashes and persists the new password,
// sends a success notification, then burns the token so it cannot be reused.
func (s *userService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	// Resolve the token to the owner's email address stored in Redis.
	storedEmail, err := s.tokenRepo.GetResetPasswordToken(ctx, req.Token)
	if err != nil {
		return pkg.ErrInvalidOrExpiredToken
	}

	// Fetch the user record for name personalisation in the success email.
	user, err := s.repo.GetUserByEmail(ctx, storedEmail)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrNotFound
		}
		return err
	}

	// Hash the incoming password before writing it to the database.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return pkg.ErrInternal
	}

	if err := s.repo.UpdatePassword(ctx, &domain.UpdatePasswordEntity{
		Email:    user.Email,
		Password: string(hashedPassword),
	}); err != nil {
		return err
	}

	// Safely dereference the optional Name pointer to avoid a nil-panic.
	userName := ""
	if user.Name != nil {
		userName = *user.Name
	}

	// Notify the user that their password was changed — important for security awareness.
	s.sendEmailJob(ctx, EmailJob{
		Type: EmailJobUpdatePasswordSuccess,
		To:   user.Email,
		Name: userName,
	})

	// Burn the token — it's single-use and must not be reusable after a successful reset.
	_ = s.tokenRepo.DeleteResetPasswordToken(ctx, req.Token)

	return nil
}

// ==== User CRUD ====

// GetUsers returns all registered users. Intended for admin use only.
func (s *userService) GetUsers(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return ToUserListResponse(users), nil
}

// GetUserByID fetches a single user by their UUID.
func (s *userService) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ToUserResponse(user), nil
}

// UpdateUser allows an admin to update any user's fields by ID.
// The cache is synced after a successful write to keep Redis consistent.
func (s *userService) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UserResponse, error) {
	*req.Email = strings.ToLower(*req.Email)
	
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	update := ToUpdateUserRequest(req)
	if err := s.repo.UpdateUser(ctx, update); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Keep Redis in sync with the latest state from the DB.
	_ = s.cacheRepo.SetUserCache(ctx, ToUserCache(user))

	return ToUserResponse(user), nil
}

// UpdateMyProfile lets an authenticated user update their own non-sensitive fields.
// Role and email changes are intentionally excluded here — those go through dedicated flows.
func (s *userService) UpdateMyProfile(ctx context.Context, req *UpdateMySelfRequest) (*UserResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	update := ToUpdateUserRequest(&UpdateUserRequest{
		ID:       req.ID,
		Name:     req.Name,
		Username: req.Username,
		Avatar:   req.Avatar,
	})
	if err := s.repo.UpdateUser(ctx, update); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Reflect the update in Redis so middleware cache reads stay accurate.
	_ = s.cacheRepo.SetUserCache(ctx, ToUserCache(user))

	return ToUserResponse(user), nil
}

// CompleteProfile handles the post-registration profile setup step where users pick a username
// and fill in optional details. It enforces global username uniqueness before writing.
func (s *userService) CompleteProfile(ctx context.Context, req CompleteProfileRequest) (*UserResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	// Check for username conflicts, but allow the current user to keep their existing one.
	existing, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, pkg.ErrNotFound) {
		return nil, err
	}
	if existing != nil && existing.ID != req.UserID {
		return nil, pkg.ErrUsernameAlreadyUsed
	}

	// Stamp profile_completed_at to mark the onboarding step as done.
	now := time.Now()
	update := ToCompleteProfileRequest(req)
	update.ProfileCompletedAt = &now

	if err := s.repo.UpdateUser(ctx, update); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Sync the cache so the middleware sees the updated profile_completed_at flag.
	_ = s.cacheRepo.SetUserCache(ctx, ToUserCache(user))

	return ToUserResponse(user), nil
}

// UpdatePassword lets an authenticated user change their password after verifying the current one.
// A success notification email is sent once the new password is stored.
func (s *userService) UpdatePassword(ctx context.Context, req ChangePassword) error {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return err
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrNotFound
		}
		return err
	}

	// Verify the current password before allowing any change.
	oldPassword, _, err := s.repo.GetPasswordByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.ErrInvalidPassword
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(req.OldPassword)); err != nil {
		return pkg.ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return pkg.ErrInternal
	}

	if err := s.repo.UpdatePassword(ctx, &domain.UpdatePasswordEntity{
		Email:    req.Email,
		Password: string(hashedPassword),
	}); err != nil {
		return err
	}

	// Safely dereference optional Name pointer before passing it to the email job.
	userName := ""
	if user.Name != nil {
		userName = *user.Name
	}

	// Alert the user about the password change — critical for account security.
	s.sendEmailJob(ctx, EmailJob{
		Type: EmailJobUpdatePasswordSuccess,
		To:   user.Email,
		Name: userName,
	})

	return nil
}

// DeleteUser hard-deletes a user by ID (admin action) and evicts their session and cache entry.
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return err
	}

	// Clean up associated state so stale data doesn't linger in Redis.
	_ = s.cacheRepo.DeleteUserCache(ctx, id)
	_ = s.jwtRepo.RevokeUserSession(ctx, id)

	return nil
}

// DeleteAccount lets a user permanently remove their own account and immediately
// invalidates their active tokens to prevent any further authenticated requests.
func (s *userService) DeleteAccount(ctx context.Context, id, accessToken, refreshToken string) error {
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return err
	}

	// Revoke only the current token pair — not the entire session — since the user
	// may have other devices they weren't aware of (handled by separate security flows).
	_ = s.cacheRepo.DeleteUserCache(ctx, id)
	_ = s.jwtRepo.RevokeTokens(ctx, accessToken, refreshToken)

	return nil
}

// ==== User Address ====

func (s *userService) AddAddress(ctx context.Context, req AddAddressRequest) (*UserAddressResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	ID := uuid.NewString()

	addAdressPayload := ToAddAddressRequest(ID, &req)

	addAddress, err := s.repo.CreateAddress(ctx, addAdressPayload)
	if err != nil {
		return nil, err
	}

	return ToUserAddressResponse(addAddress), nil
}

func (s *userService) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*UserAddressResponse, error) {
	if err := jsonValidator.ValidateStruct(s.validate, req); err != nil {
		return nil, err
	}

	update := ToUpdateAddressRequest(req)
	if err := s.repo.UpdateAddress(ctx, update); err != nil {
		return nil, err
	}

	address, err := s.repo.GetAddressByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return ToUserAddressResponse(address), nil
}

func (s *userService) GetAddresses(ctx context.Context) ([]UserAddressResponse, error) {
	addresses, err := s.repo.GetAddresses(ctx)
	if err != nil {
		return nil, err
	}

	return ToUserAddressListResponse(addresses), nil
}

func (s *userService) GetAddressByID(ctx context.Context, id string) (*UserAddressResponse, error) {
	address, err := s.repo.GetAddressByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ToUserAddressResponse(address), nil
}

func (s *userService) GetAddressesByUserID(ctx context.Context, id string) ([]UserAddressResponse, error) {
	addresses, err := s.repo.GetAddressesByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ToUserAddressListResponse(addresses), nil
}

func (s *userService) DeleteAddress(ctx context.Context, addressID, userID string) error {
	if err := s.repo.DeleteAddress(ctx, addressID, userID); err != nil {
		return err
	}

	return nil
}

// ==== Internal Helpers ====

// sendRegistrationToken generates a cryptographically secure 24-hour verification token,
// stores it in Redis keyed by the token value, and enqueues the verification email job.
func (s *userService) sendRegistrationToken(ctx context.Context, userID, email string) error {
	token, err := generateSecureToken()
	if err != nil {
		return pkg.ErrInternal
	}

	if err := s.tokenRepo.SetRegisterToken(ctx, token, userID, 24*time.Hour); err != nil {
		return err
	}

	return s.sendEmailJob(ctx, EmailJob{
		Type:  EmailJobVerifyRegister,
		To:    email,
		Token: token,
	})
}

// sendEmailJob is the central dispatcher for all outbound email jobs.
// Wrapping producer calls here keeps individual service methods clean and
// makes it easy to swap the underlying broker without touching business logic.
func (s *userService) sendEmailJob(ctx context.Context, job EmailJob) error {
	if err := s.producer.PublishEmailJob(ctx, job); err != nil {
		return err
	}
	return nil
}

// generateSecureToken produces a URL-safe, cryptographically random 64-character hex string
// backed by 32 bytes from crypto/rand — safe for use as one-time tokens.
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}