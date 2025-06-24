package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spurge/p4rsec/server/internal/database"
)

type HealthHandler struct {
	db    *database.PostgresDB
	redis *database.RedisDB
}

func NewHealthHandler(db *database.PostgresDB, redis *database.RedisDB) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
}

type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceInfo),
	}

	// Check PostgreSQL
	if err := h.db.Health(ctx); err != nil {
		response.Services["postgresql"] = ServiceInfo{
			Status:  "error",
			Message: err.Error(),
		}
		response.Status = "error"
	} else {
		response.Services["postgresql"] = ServiceInfo{
			Status: "ok",
		}
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		response.Services["redis"] = ServiceInfo{
			Status:  "error",
			Message: err.Error(),
		}
		response.Status = "error"
	} else {
		response.Services["redis"] = ServiceInfo{
			Status: "ok",
		}
	}

	// Return appropriate status code
	if response.Status == "error" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}
