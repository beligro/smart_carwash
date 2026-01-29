package handlers

import (
	userService "carwash_backend/internal/domain/user/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userService userService.Service
}

func NewHandler(userService userService.Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) GetLoyaltyProgress(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user_id"})
		return
	}

	count, nextFree, err := h.userService.GetLoyaltyProgress(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"count":             count,
		"next_free_at":      nextFree,
		"is_free_available": count > 0 && count%10 == 0,
	})
}
