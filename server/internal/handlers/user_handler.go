package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/spurge/p4rsec/server/internal/dao"
	"github.com/spurge/p4rsec/server/internal/logger"
	"github.com/spurge/p4rsec/server/internal/models"
)

type UserHandler struct {
	userDAO  *dao.UserDAO
	cacheDAO *dao.CacheDAO
	logger   *logger.Logger
}

func NewUserHandler(userDAO *dao.UserDAO, cacheDAO *dao.CacheDAO, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userDAO:  userDAO,
		cacheDAO: cacheDAO,
		logger:   logger,
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Try to get from cache first
	if users, err := h.cacheDAO.GetUsers(ctx, page, limit); err == nil {
		h.logger.Debug("Users retrieved from cache", "page", page, "limit", limit)
		return c.JSON(fiber.Map{
			"users": users,
			"page":  page,
			"limit": limit,
		})
	}

	// Get from database
	users, err := h.userDAO.GetAll(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to retrieve users",
		})
	}

	// Cache the result
	if err := h.cacheDAO.SetUsers(ctx, users, page, limit); err != nil {
		h.logger.Warn("Failed to cache users", "error", err)
	}

	return c.JSON(fiber.Map{
		"users": users,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid user ID format",
		})
	}

	// Try cache first
	if user, err := h.cacheDAO.GetUser(ctx, userID.String()); err == nil {
		h.logger.Debug("User retrieved from cache", "user_id", userID)
		return c.JSON(fiber.Map{
			"user": user,
		})
	}

	// Get from database
	user, err := h.userDAO.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "user_id", userID)
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to retrieve user",
		})
	}

	// Cache the result
	if err := h.cacheDAO.SetUser(ctx, user); err != nil {
		h.logger.Warn("Failed to cache user", "error", err, "user_id", userID)
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Basic validation
	if req.Email == "" || req.Username == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "All fields are required",
		})
	}

	// Check if user already exists
	if existingUser, err := h.userDAO.GetByEmail(ctx, req.Email); err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   true,
			"message": "User with this email already exists",
		})
	}

	// Create user
	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.userDAO.Create(ctx, user); err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create user",
		})
	}

	// Cache the new user
	if err := h.cacheDAO.SetUser(ctx, user); err != nil {
		h.logger.Warn("Failed to cache new user", "error", err, "user_id", user.ID)
	}

	// Invalidate users list cache
	if err := h.cacheDAO.InvalidateUsersList(ctx); err != nil {
		h.logger.Warn("Failed to invalidate users list cache", "error", err)
	}

	h.logger.Info("User created successfully", "user_id", user.ID, "email", user.Email)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid user ID format",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "No updates provided",
		})
	}

	// Update user
	if err := h.userDAO.Update(ctx, userID, updates); err != nil {
		h.logger.Error("Failed to update user", "error", err, "user_id", userID)
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update user",
		})
	}

	// Invalidate cache
	if err := h.cacheDAO.DeleteUser(ctx, userID.String()); err != nil {
		h.logger.Warn("Failed to invalidate user cache", "error", err, "user_id", userID)
	}
	if err := h.cacheDAO.InvalidateUsersList(ctx); err != nil {
		h.logger.Warn("Failed to invalidate users list cache", "error", err)
	}

	// Get updated user
	user, err := h.userDAO.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get updated user", "error", err, "user_id", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "User updated but failed to retrieve updated data",
		})
	}

	// Cache the updated user
	if err := h.cacheDAO.SetUser(ctx, user); err != nil {
		h.logger.Warn("Failed to cache updated user", "error", err, "user_id", userID)
	}

	h.logger.Info("User updated successfully", "user_id", userID)

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := c.Params("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid user ID format",
		})
	}

	// Delete user (soft delete)
	if err := h.userDAO.Delete(ctx, userID); err != nil {
		h.logger.Error("Failed to delete user", "error", err, "user_id", userID)
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete user",
		})
	}

	// Invalidate cache
	if err := h.cacheDAO.DeleteUser(ctx, userID.String()); err != nil {
		h.logger.Warn("Failed to invalidate user cache", "error", err, "user_id", userID)
	}
	if err := h.cacheDAO.InvalidateUsersList(ctx); err != nil {
		h.logger.Warn("Failed to invalidate users list cache", "error", err)
	}

	h.logger.Info("User deleted successfully", "user_id", userID)

	return c.Status(fiber.StatusNoContent).Send(nil)
}
