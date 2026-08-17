package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"E-COMMERCE-API/internal/categories"
	"E-COMMERCE-API/internal/config"
	"E-COMMERCE-API/internal/infrastructure/postgres"
	"E-COMMERCE-API/internal/infrastructure/redpanda"
	"E-COMMERCE-API/internal/jwt"
	"E-COMMERCE-API/internal/middleware"
	"E-COMMERCE-API/internal/products"
	"E-COMMERCE-API/internal/routes"
	"E-COMMERCE-API/internal/stores"
	"E-COMMERCE-API/internal/users"
	"E-COMMERCE-API/pkg"

	redisinfra "E-COMMERCE-API/internal/infrastructure/redis"
)

func main() {
	cfg      := config.NewConfig()
	validate := pkg.New()
	logger   := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// ── Infrastructure ────────────────────────────────────────────────────────
	redisClient := redisinfra.NewRedisClient(cfg.REDISaddr, cfg.REDISpassword)

	db, err := postgres.NewPostgresDB(cfg.DSN)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}

	// ── Redpanda ──────────────────────────────────────────────────────────────
	if err := redpanda.EnsureTopics(cfg.REDPANDAbrokers, "email-jobs"); err != nil {
		log.Fatalf("failed to ensure redpanda topics: %v", err)
	}

	emailProducer, err := users.NewRedpandaEmailProducer(cfg.REDPANDAbrokers)
	if err != nil {
		log.Fatalf("failed to create email producer: %v", err)
	}
	defer emailProducer.Close()

	// ── Email Worker (consumer) ───────────────────────────────────────────────
	emailSvc := users.NewEmailService(cfg)

	emailWorker, err := users.NewEmailWorker(cfg.REDPANDAbrokers, "email-worker-group", cfg.APPurl, emailSvc, logger)
	if err != nil {
		log.Fatalf("failed to create email worker: %v", err)
	}
	defer emailWorker.Close()

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go emailWorker.Run(workerCtx)

	// ── Services ──────────────────────────────────────────────────────────────
	jwtService := jwt.NewJWTService(cfg.JWTSecretKey, cfg.JWTExpiry, cfg.DefaultRefreshExpiry, cfg.ShortRefreshExpiry, redisClient)

	userRepo      := users.NewUserRepository(db)
	userTokenRepo := users.NewUserTokenRepository(redisClient)
	userCacheRepo := users.NewUserCacheRepository(redisClient)
	storeRepo     := stores.NewStoreRepository(db)
	ctgryRepo     := categories.NewCategoryRepository(db)
	prodctRepo    := products.NewProductRepository(db)

	userService   := users.NewUserService(userRepo, userTokenRepo, jwtService, emailProducer, userCacheRepo, validate)
	storeService  := stores.NewStoreService(storeRepo, userRepo, userCacheRepo, validate)
	ctgryService  := categories.NewCategoryService(ctgryRepo, validate)
	prodctService := products.NewProductService(prodctRepo, storeRepo, validate)

	// ── Handlers ──────────────────────────────────────────────────────────────
	userHandler    := users.NewUserHandler(userService, cfg.DefaultRefreshExpiry, cfg.ShortRefreshExpiry)
	storeHandler   := stores.NewStoreHandler(storeService)
	ctgryHandler   := categories.NewCategoryHandler(ctgryService)
	prodctHandler  := products.NewProductHandler(prodctService)
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userRepo, userCacheRepo)

	r := routes.NewUserRouter(userHandler, storeHandler, ctgryHandler, prodctHandler, authMiddleware)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.APPport,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("Server running on :%s", cfg.APPport)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop worker dulu sebelum tutup koneksi
	workerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("DB close error: %v", err)
		}
	}

	log.Println("Server exited")
}