package repository

import (
	"context"
	"sub_service/internal/models"

	"github.com/google/uuid"
)

// SubscriptionRepo - интерфейс для работы с подписками
type SubscriptionRepo interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, update *models.SubscriptionUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*models.Subscription, error)
	CalculateTotalCost(ctx context.Context, req *models.TotalCostRequest) (int, error)
}
