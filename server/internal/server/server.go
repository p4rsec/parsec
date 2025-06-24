package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spurge/p4rsec/server/internal/config"
	"github.com/spurge/p4rsec/server/internal/dao"
	"github.com/spurge/p4rsec/server/internal/database"
	"github.com/spurge/p4rsec/server/internal/handlers"
	appLogger "github.com/spurge/p4rsec/server/internal/logger"
)

type Server struct {
	app    *fiber.App
	config *config.Config
	logger *appLogger.Logger
	db     *database.PostgresDB
	redis  *database.RedisDB
}

func New(cfg *config.Config, logger *appLogger.Logger, db *database.PostgresDB, redis *database.RedisDB) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error("Request error", "error", err, "path", c.Path(), "method", c.Method())
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	server := &Server{
		app:    app,
		config: cfg,
		logger: logger,
		db:     db,
		redis:  redis,
	}

	server.setupMiddlewares()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddlewares() {
	// Recover from panics
	s.app.Use(recover.New())

	// Security headers
	s.app.Use(helmet.New())

	// CORS
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Rate limiting
	s.app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Rate limit exceeded",
			})
		},
	}))

	// Request logging
	if s.config.Environment != "production" {
		s.app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} ${latency}\n",
		}))
	}
}

func (s *Server) setupRoutes() {
	// Initialize DAOs
	userDAO := dao.NewUserDAO(s.db)
	cacheDAO := dao.NewCacheDAO(s.redis)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(s.db, s.redis)
	userHandler := handlers.NewUserHandler(userDAO, cacheDAO, s.logger)

	// API routes
	api := s.app.Group("/api/v1")

	// Health check
	api.Get("/health", healthHandler.Health)

	// User routes
	users := api.Group("/users")
	users.Get("/", userHandler.GetUsers)
	users.Post("/", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Root route
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":     "P4rsec API Server",
			"version":     "1.0.0",
			"environment": s.config.Environment,
		})
	})

	// 404 handler
	s.app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Route not found",
		})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
