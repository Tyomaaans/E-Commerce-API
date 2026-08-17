package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	category "E-COMMERCE-API/internal/categories"
	"E-COMMERCE-API/internal/domain"
	"E-COMMERCE-API/internal/middleware"
	product "E-COMMERCE-API/internal/products"
	store "E-COMMERCE-API/internal/stores"
	user "E-COMMERCE-API/internal/users"
)

func NewUserRouter(
	userHandler     *user.UserHandler,
	storeHandler    *store.StoreHandler,
	categoryHandler *category.CategoryHandler,
	productHandler  *product.ProductHandler,
	authMiddleware  *middleware.AuthMiddleware,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/api/v1")
	{
		// AUTH
		auth := v1.Group("/auth")
		{
			auth.POST("/register",            userHandler.Register)
			auth.POST("/resend-verification", userHandler.ResendVerification)
			auth.GET("/verify-email",         userHandler.VerifyRegister)
			auth.POST("/login",               userHandler.Login)
			auth.POST("/refresh",             userHandler.RefreshToken)
			auth.POST("/forgot-password",     userHandler.ForgotPassword)
			auth.GET("/reset-password",       userHandler.VerifyResetPassword)
			auth.POST("/reset-password",      userHandler.ResetPassword)

			// Requires auth
			auth.POST("/logout", authMiddleware.Authenticate(), userHandler.Logout)
		}

		// USER (authenticated + verified)
		users := v1.Group("/users")
		users.Use(
			authMiddleware.Authenticate(), 
			authMiddleware.RequireVerifiedEmail(),
		)
		{
			me := users.Group("/me")
			{
				// Profile setup (tidak butuh completed profile)
				me.PUT("/profile", userHandler.CompleteProfile)
				me.DELETE("",      userHandler.DeleteAccount)

				// Requires completed profile
				meProfile := me.Group("")
				meProfile.Use(authMiddleware.RequireCompletedProfile())
				{
					meProfile.GET("",          userHandler.GetMyProfile)
					meProfile.PATCH("",        userHandler.UpdateMyProfile)
					meProfile.PUT("/password", userHandler.ChangePassword)

					// Addresses
					meProfile.GET("/addresses",        userHandler.GetAddressesByUserID)
					meProfile.POST("/addresses",       userHandler.AddAddress)
					meProfile.PATCH("/addresses/:id",  userHandler.UpdateAddress)
					meProfile.DELETE("/addresses/:id", userHandler.DeleteAddress)
				}
			}
		}

		// STORES
		stores := v1.Group("/stores")
		{
			// Public
			stores.GET("",     storeHandler.GetStores)
			stores.GET("/:id", storeHandler.GetStoreByID)

			// Buyer: register store
			buyerStores := stores.Group("")
			buyerStores.Use(
				authMiddleware.Authenticate(),
				authMiddleware.RequireVerifiedEmail(),
				authMiddleware.RequireCompletedProfile(),
				authMiddleware.RequireRole(string(domain.Buyer)),
			)
			{
				buyerStores.POST("", storeHandler.RegisterStore)
			}

			// Seller: own store
			sellerStores := stores.Group("/me")
			sellerStores.Use(
				authMiddleware.Authenticate(),
				authMiddleware.RequireVerifiedEmail(),
				authMiddleware.RequireCompletedProfile(),
				authMiddleware.RequireRole(string(domain.Seller)),
			)
			{
				sellerStores.GET("",    storeHandler.GetStoreByUserID)
				sellerStores.PATCH("",  storeHandler.UpdateStore)
				sellerStores.DELETE("", storeHandler.DeactivateStore)

				// Seller: own products
				sellerStores.POST("/products",                productHandler.AddProduct)
				sellerStores.PATCH("/products/:id",           productHandler.UpdateProduct)
				sellerStores.POST("/products/:id/deactivate", productHandler.DeactivateProduct)
				sellerStores.POST("/products/:id/reactivate", productHandler.ReactivateProduct)
			}
		}

		// CATEGORIES
		categories := v1.Group("/categories")
		{
			// Public
			categories.GET("",     categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategoryByID)
		}

		// PRODUCTS
		products := v1.Group("/products")
		{
			// Public
			products.GET("",     productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProductByID)
		}

		// ADMIN
		admin := v1.Group("/admin")
		admin.Use(
			authMiddleware.Authenticate(), 
			authMiddleware.RequireRole(string(domain.Admin)),
		)
		{
			// Users
			admin.GET("/users",        userHandler.GetUsers)
			admin.GET("/users/:id",    userHandler.GetUserByID)
			admin.PATCH("/users/:id",  userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)

			// Addresses
			admin.GET("/addresses",                      userHandler.GetAddresses)
			admin.GET("/addresses/:id",                  userHandler.GetAddressByID)
			admin.PATCH("/users/:userID/addresses/:id",  userHandler.UpdateUserAddress)
			admin.DELETE("/users/:userID/addresses/:id", userHandler.DeleteUserAddress)

			// Stores
			admin.GET("/stores",                 storeHandler.GetInactiveStores)
			admin.GET("/stores/:userID",         storeHandler.GetUserStoreByUserID)
			admin.PATCH("/stores/:id",           storeHandler.UpdateUserStore)
			admin.POST("/stores/:id/reactivate", storeHandler.ReactivateUserStore)
			admin.POST("/stores/:id/deactivate", storeHandler.DeactivateUserStore) // fix: DELETE → POST/PATCH lebih RESTful untuk soft action
			admin.DELETE("/stores/:id",          storeHandler.DeleteUserStore)

			// Categories
			admin.POST("/categories",                   categoryHandler.AddParentCategory)
			admin.POST("/categories/:id/subcategories", categoryHandler.AddChildCategory)
			admin.PATCH("/categories/:id",              categoryHandler.UpdateCategory)
			admin.PUT("/categories/:id/parent",         categoryHandler.ChangeParentCategory)
			admin.DELETE("/categories/:id",             categoryHandler.DeleteCategory)

			// Products
			admin.PATCH("/users/:userID/products/:id",           productHandler.AdminUpdateProduct)
			admin.POST("/users/:userID/products/:id/deactivate", productHandler.AdminDeactivateProduct) // fix: DELETE → POST untuk soft action
			admin.POST("/users/:userID/products/:id/reactivate", productHandler.AdminReactivateProduct)
			admin.DELETE("/users/:userID/products/:id",          productHandler.DeleteProduct)
		}
	}

	return r
}

/*
package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"E-COMMERCE-API/internal/middleware"
	user "E-COMMERCE-API/internal/users"
)

func NewUserRouter(
	userHandler    *user.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	r := gin.Default()

	// Allow requests from the frontend dev server.
	// Swap AllowOrigins for an env-based list before going to production.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/api/v1")

	registerAuthRoutes(v1, userHandler, authMiddleware)
	registerUserRoutes(v1, userHandler, authMiddleware)

	return r
}

// registerAuthRoutes mounts public-facing auth endpoints under /api/v1/auth.
// Logout is the only route here that requires a valid token — it needs one
// to know which session to invalidate.
func registerAuthRoutes(
	rg *gin.RouterGroup,
	h  *user.UserHandler,
	am *middleware.AuthMiddleware,
) {
	auth := rg.Group("/auth")

	auth.POST("/register",              h.Register)
	auth.POST("/resend-verification",   h.ResendVerification)
	auth.GET("/verify-email",           h.VerifyRegister)
	auth.POST("/login",                 h.Login)
	auth.POST("/refresh",               h.RefreshToken)
	auth.POST("/logout",                am.Authenticate(), h.Logout)
	auth.POST("/forgot-password",       h.ForgotPassword)
	auth.GET("/reset-password",         h.VerifyResetPassword)
	auth.POST("/reset-password",        h.ResetPassword)
}

// registerUserRoutes mounts routes that require a verified account.
// Every endpoint under /api/v1/users passes through Authenticate and
// RequireVerifiedEmail first; profile-sensitive ones also check
// RequireCompletedProfile.
func registerUserRoutes(
	rg *gin.RouterGroup,
	h  *user.UserHandler,
	am *middleware.AuthMiddleware,
) {
	users := rg.Group("/users")
	users.Use(am.Authenticate(), am.RequireVerifiedEmail())

	// Profile — reading and editing require a complete profile,
	// but completing it obviously can't be behind that same gate.
	mustComplete := am.RequireCompletedProfile()
	users.GET("/me",                  mustComplete, h.GetMyProfile)
	users.PATCH("/me",                mustComplete, h.UpdateMyProfile)
	users.PATCH("/me/avatar",         mustComplete, h.UpdateAvatar)          // upload/replace profile picture
	users.POST("/me/change-password", mustComplete, h.ChangePassword)        // for users who registered with email
	users.POST("/me/complete-profile", h.CompleteProfile)

	// Admin — all routes below this point are restricted to the Admin role.
	admin := users.Group("/admin")
	admin.Use(am.RequireRole(string(user.Admin)))

	admin.GET("",       h.GetUsers)     // list with pagination & filters
	admin.GET("/:id",   h.GetUserByID)
	admin.PATCH("/:id", h.UpdateUser)
	admin.DELETE("/:id", h.DeleteUser)
	admin.PATCH("/:id/ban",    h.BanUser)    // soft-ban; keeps the record, blocks login
	admin.PATCH("/:id/unban",  h.UnbanUser)
}
*/