// @Title Subscription Service API
// @Version 1.0
// @Description API для управления подписками пользователей

// @Contact.name API Support
// @Contact.email rodnynskiy@gmail.com

// @Host localhost:8080
// @BasePath /api/v1

package handler

import (
	"net/http"

	"sub_service/internal/models"
	"sub_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *zap.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{service: service, logger: logger}
}

// CreateSubscription godoc
// @Summary Создать новую подписку
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param input body models.SubscriptionCreate true "Данные подписки"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var create models.SubscriptionCreate
	if err := c.ShouldBindJSON(&create); err != nil {
		h.logger.Warn("Invalid create request",
			zap.Error(err),
			zap.Any("request", c.Request.Body),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Handling create subscription request",
		zap.String("service_name", create.ServiceName),
		zap.String("user_id", create.UserID.String()),
	)

	sub, err := h.service.Create(c.Request.Context(), &create)
	if err != nil {
		h.logger.Error("Failed to create subscription",
			zap.Error(err),
			zap.Any("request", create),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Subscription created successfully",
		zap.String("id", sub.ID.String()),
	)
	c.JSON(http.StatusCreated, sub)
}

// GetSubscriptionByID godoc
// @Summary Получить подписку по ID
// @Tags Subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]

func (h *SubscriptionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Invalid subscription ID format",
			zap.String("id_param", c.Param("id")),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	h.logger.Debug("Handling get subscription request",
		zap.String("id", id.String()),
	)

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sub == nil {
		h.logger.Warn("Subscription not found",
			zap.String("id", id.String()),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	h.logger.Debug("Subscription retrieved",
		zap.String("id", id.String()),
	)
	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Tags Subscriptions
// @Accept json
// @Param id path string true "ID подписки"
// @Param input body models.SubscriptionUpdate true "Обновляемые данные"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]

func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Invalid subscription ID format",
			zap.String("id_param", c.Param("id")),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var update models.SubscriptionUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Warn("Invalid update request",
			zap.Error(err),
			zap.Any("request", c.Request.Body),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Handling update subscription request",
		zap.String("id", id.String()),
		zap.Any("update", update),
	)

	if err := h.service.Update(c.Request.Context(), id, &update); err != nil {
		h.logger.Error("Failed to update subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Subscription updated successfully",
		zap.String("id", id.String()),
	)
	c.Status(http.StatusOK)
}

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Tags Subscriptions
// @Param id path string true "ID подписки"
// @Success 204
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Invalid subscription ID format",
			zap.String("id_param", c.Param("id")),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	h.logger.Info("Handling delete subscription request",
		zap.String("id", id.String()),
	)

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Subscription deleted successfully",
		zap.String("id", id.String()),
	)
	c.Status(http.StatusNoContent)
}

// GetSubscriptions godoc
// @Summary Получить список всех подписок
// @Tags Subscriptions
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]

func (h *SubscriptionHandler) List(c *gin.Context) {
	h.logger.Debug("Handling list subscriptions request")

	subs, err := h.service.List(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list subscriptions",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Subscriptions listed",
		zap.Int("count", len(subs)),
	)
	c.JSON(http.StatusOK, subs)
}

// CalculateTotalCost godoc
// @Summary Рассчитать суммарную стоимость
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param input body models.TotalCostRequest true "Параметры расчета"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total-cost [post]

func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	var req models.TotalCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid total cost request",
			zap.Error(err),
			zap.Any("request", c.Request.Body),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Handling calculate total cost request",
		zap.Any("request", req),
	)

	total, err := h.service.CalculateTotalCost(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to calculate total cost",
			zap.Error(err),
			zap.Any("request", req),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Total cost calculated",
		zap.Int("total", total),
		zap.Any("request", req),
	)
	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: total})
}
