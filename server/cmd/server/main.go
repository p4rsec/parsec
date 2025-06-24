package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spurge/p4rsec/server/internal/config"
	"github.com/spurge/p4rsec/server/internal/database"
	"github.com/spurge/p4rsec/server/internal/logger"
	"github.com/spurge/p4rsec/server/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Logger.Level, cfg.Environment)
	defer logger.Sync()

	// Initialize database connections
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}

	redis, err := database.NewRedisConnection(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}

	// Initialize and start server
	srv := server.New(cfg, logger, db, redis)
	
	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	logger.Info("Server started", "port", cfg.Server.Port, "environment", cfg.Environment)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server gracefully", "error", err)
	}

	// Close database connections
	if err := db.Close(); err != nil {
		logger.Error("Failed to close database connection", "error", err)
	}

	if err := redis.Close(); err != nil {
		logger.Error("Failed to close Redis connection", "error", err)
	}

	logger.Info("Server stopped")
} 