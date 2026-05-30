package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/sfedu-crm/internal/config"
	httphandler "github.com/sfedu-crm/internal/handler/http"
	"github.com/sfedu-crm/internal/middleware"
	pgrepo "github.com/sfedu-crm/internal/repository/postgres"
	"github.com/sfedu-crm/internal/service"
	rediscache "github.com/sfedu-crm/pkg/cache"
	"github.com/sfedu-crm/pkg/hash"
	"github.com/sfedu-crm/pkg/jwt"
	"github.com/sfedu-crm/pkg/logger"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	log := logger.New(os.Getenv("APP_ENV"))

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ---- БД ----
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("db ping failed", "error", err)
		os.Exit(1)
	}
	log.Info("connected to database")

	// Redis cache
	cache := rediscache.NewRedisCache(cfg.RedisAddr)
	if err := cache.Ping(context.Background()); err != nil {
		log.Warn("redis not available, running without cache", "err", err)
	} else {
		log.Info("redis connected")
	}

	// ---- Миграции ----
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to create migrator", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied")

	// ---- Обновляем хеш пароля админа из env ----
	_, err = pool.Exec(ctx,
		`UPDATE users SET password_hash=$1 WHERE email='admin@gmail.com' AND password_hash='PLACEHOLDER'`,
		cfg.AdminPassword,
	)
	if err != nil {
		log.Error("failed to seed admin password", "error", err)
		os.Exit(1)
	}

	// ---- Инфраструктура ----
	hasher := hash.NewBcrypt(cfg.BcryptCost)
	tokenMgr := jwt.NewManager(cfg.JWTSecret)

	// ---- Репозитории ----
	userRepo := pgrepo.NewUserRepository(pool)
	appRepo := pgrepo.NewApplicationRepository(pool)
	trainerRepo := pgrepo.NewTrainerRepository(pool)
	trainingRepo := pgrepo.NewTrainingRepository(pool)
	trainingReqRepo := pgrepo.NewTrainingRequestRepository(pool)
	productRepo := pgrepo.NewProductRepository(pool)
	subscriptionRepo := pgrepo.NewSubscriptionRepository(pool)
	paymentRepo := pgrepo.NewPaymentRepository(pool)

	// ---- Сервисы ----
	authSvc := service.NewAuthService(userRepo, appRepo, tokenMgr, hasher, cfg.JWTTokenTTL)
	userSvc := service.NewUserService(userRepo, appRepo, hasher, cache)
	trainerSvc := service.NewTrainerService(trainerRepo, cache)
	scheduleSvc := service.NewScheduleService(trainingRepo, trainingReqRepo, userRepo, paymentRepo, subscriptionRepo)
	financeSvc := service.NewFinanceService(paymentRepo, userRepo)
	shopSvc := service.NewShopService(productRepo, subscriptionRepo, paymentRepo, userRepo, cache)
	aiSvc := service.NewAIAssistantService()

	// ---- Хэндлеры ----
	handlers := &httphandler.Handlers{
		Auth:     httphandler.NewAuthHandler(authSvc),
		User:     httphandler.NewUserHandler(userSvc),
		Schedule: httphandler.NewScheduleHandler(scheduleSvc),
		Finance:  httphandler.NewFinanceHandler(financeSvc),
		Shop:     httphandler.NewShopHandler(shopSvc),
		Trainer:  httphandler.NewTrainerHandler(trainerSvc),
		AI:       httphandler.NewAIHandler(aiSvc),
	}

	// ---- Роутер с глобальными middleware ----
	router := httphandler.NewRouter(handlers, tokenMgr)
	handler := middleware.Logging(log)(middleware.CORS(cfg.CORSOrigin)(router))

	// ---- HTTP сервер ----
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- Graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced shutdown", "error", err)
	}
	log.Info("server exited")
}
