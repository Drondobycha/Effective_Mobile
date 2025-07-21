package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sub_service/internal/config"
	"sub_service/internal/handler"
	"sub_service/internal/repository/postgres"
	"sub_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}
	defer logger.Sync()

	// Инициализация конфигурации
	cfg := config.MustLoad()

	// Подключение к базе данных
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Error connecting to database", zap.Error(err))
	}
	defer db.Close()

	// Инициализация слоев приложения
	repo := postgres.NewSubscriptionRepo(db, logger)
	subscriptionService := service.NewSubscriptionService(repo, logger)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, logger)

	// Настройка роутера
	router := gin.Default()

	// Добавление middleware для логирования запросов
	router.Use(func(c *gin.Context) {
		logger.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)
		c.Next()
	})

	api := router.Group("/api/v1")
	{
		api.POST("/subscriptions", subscriptionHandler.Create)
		api.GET("/subscriptions", subscriptionHandler.List)
		api.GET("/subscriptions/:id", subscriptionHandler.Get)
		api.PUT("/subscriptions/:id", subscriptionHandler.Update)
		api.DELETE("/subscriptions/:id", subscriptionHandler.Delete)
		api.POST("/subscriptions/total-cost", subscriptionHandler.CalculateTotalCost)
	}

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting server", zap.Error(err))
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server exited properly")
}
