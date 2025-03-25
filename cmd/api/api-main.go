package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"admin-statistics-api/internal/config"
	"admin-statistics-api/internal/handler"
	"admin-statistics-api/internal/middleware"
	"admin-statistics-api/internal/repository"
	"admin-statistics-api/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Check connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")

	// Initialize repositories, services, and handlers
	db := client.Database(cfg.MongoDB.Database)
	transactionRepo := repository.NewTransactionRepository(db, cfg.MongoDB.Collection)
	
	// Initialize Redis cache
	redisCache, err := repository.NewRedisCache(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()
	
	log.Println("Connected to Redis successfully")
	
	// Ensure we're using the correct interface type
	var cache repository.Cache = redisCache
	
	transactionService := service.NewTransactionService(transactionRepo, cache)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	// Initialize Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.AuthMiddleware(cfg))

	// Define routes
	router.GET("/gross_gaming_rev", transactionHandler.GetGrossGamingRevenue)
	router.GET("/daily_wager_volume", transactionHandler.GetDailyWagerVolume)
	router.GET("/user/:user_id/wager_percentile", transactionHandler.GetUserWagerPercentile)

	// Start HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give the server time to shutdown gracefully
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}