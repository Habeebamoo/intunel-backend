package handlers

import (
	"net/http"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/services"
	"github.com/Habeebamoo/intunel-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service services.NotificationService
}

func NewNotificationHandler(s services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var n models.Notification

	if err := c.ShouldBindJSON(&n); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if n.Channel != "email" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid channel. Only 'email' is supported.")
		return
	}

	if err := h.service.SendNotification(c.Request.Context(), n); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusAccepted, "notification queued", nil)
}