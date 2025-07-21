package service

import (
	"context"
	"time"

	"sub_service/internal/models"
	"sub_service/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepo
	logger *zap.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepo, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func (s *SubscriptionService) Create(ctx context.Context, create *models.SubscriptionCreate) (*models.Subscription, error) {
	s.logger.Info("Creating new subscription",
		zap.String("service_name", create.ServiceName),
		zap.String("user_id", create.UserID.String()),
	)

	startDate, err := time.Parse("01-2006", create.StartDate)
	if err != nil {
		s.logger.Error("Invalid start_date format",
			zap.String("start_date", create.StartDate),
			zap.Error(err),
		)
		return nil, err
	}

	var endDate *time.Time
	if create.EndDate != nil {
		parsed, err := time.Parse("01-2006", *create.EndDate)
		if err != nil {
			s.logger.Error("Invalid end_date format",
				zap.String("end_date", *create.EndDate),
				zap.Error(err),
			)
			return nil, err
		}
		endDate = &parsed
	}

	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: create.ServiceName,
		Price:       create.Price,
		UserID:      create.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.Error("Failed to create subscription in repository",
			zap.Error(err),
			zap.Any("subscription", sub),
		)
		return nil, err
	}

	s.logger.Info("Subscription created successfully",
		zap.String("id", sub.ID.String()),
	)
	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	s.logger.Debug("Getting subscription by ID", zap.String("id", id.String()))

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get subscription from repository",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return nil, err
	}

	if sub == nil {
		s.logger.Warn("Subscription not found", zap.String("id", id.String()))
	}

	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, update *models.SubscriptionUpdate) error {
	s.logger.Info("Updating subscription",
		zap.String("id", id.String()),
		zap.Any("update", update),
	)

	if err := s.repo.Update(ctx, id, update); err != nil {
		s.logger.Error("Failed to update subscription in repository",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Subscription updated successfully",
		zap.String("id", id.String()),
	)
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting subscription", zap.String("id", id.String()))

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete subscription from repository",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Subscription deleted successfully",
		zap.String("id", id.String()),
	)
	return nil
}

func (s *SubscriptionService) List(ctx context.Context) ([]*models.Subscription, error) {
	s.logger.Debug("Listing all subscriptions")

	subs, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list subscriptions from repository",
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Debug("Retrieved subscriptions",
		zap.Int("count", len(subs)),
	)
	return subs, nil
}

func (s *SubscriptionService) CalculateTotalCost(ctx context.Context, req *models.TotalCostRequest) (int, error) {
	s.logger.Info("Calculating total cost for subscriptions",
		zap.Any("request", req),
	)

	total, err := s.repo.CalculateTotalCost(ctx, req)
	if err != nil {
		s.logger.Error("Failed to calculate total cost",
			zap.Error(err),
			zap.Any("request", req),
		)
		return 0, err
	}

	s.logger.Info("Total cost calculated",
		zap.Int("total", total),
		zap.Any("request", req),
	)
	return total, nil
}
