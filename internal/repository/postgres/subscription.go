package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sub_service/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SubscriptionRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewSubscriptionRepo создает новый экземпляр SubscriptionRepo
// @param db - подключение к PostgreSQL
// @param logger - логгер
// @return *SubscriptionRepo

func NewSubscriptionRepo(db *sqlx.DB, logger *zap.Logger) *SubscriptionRepo {
	return &SubscriptionRepo{db: db, logger: logger}
}

// Create создает новую подписку
// @Summary Создать подписку
// @Description Создает новую запись о подписке в базе данных
// @Param sub body models.Subscription true "Данные подписки"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	query := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	r.logger.Info("Creating subscription",
		zap.String("service_name", sub.ServiceName),
		zap.Int("price", sub.Price),
		zap.String("user_id", sub.UserID.String()),
	)

	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	if err != nil {
		r.logger.Error("Failed to create subscription",
			zap.Error(err),
			zap.String("service_name", sub.ServiceName),
			zap.String("user_id", sub.UserID.String()),
		)
	}
	return err
}

// GetByID получает подписку по ID
// @Summary Получить подписку
// @Description Возвращает подписку по указанному UUID
// @Param id path string true "UUID подписки"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	r.logger.Debug("Getting subscription by ID", zap.String("id", id.String()))

	var sub models.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1`
	err := r.db.GetContext(ctx, &sub, query, id)
	if err == sql.ErrNoRows {
		r.logger.Warn("Subscription not found", zap.String("id", id.String()))
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
	}
	return &sub, err
}

// Update обновляет данные подписки
// @Summary Обновить подписку
// @Description Обновляет данные существующей подписки
// @Param id path string true "UUID подписки"
// @Param update body models.SubscriptionUpdate true "Обновляемые данные"
// @Success 200
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) Update(ctx context.Context, id uuid.UUID, update *models.SubscriptionUpdate) error {
	r.logger.Info("Updating subscription",
		zap.String("id", id.String()),
		zap.Any("update", update),
	)

	query := `UPDATE subscriptions SET 
	          service_name = COALESCE($1, service_name),
	          price = COALESCE($2, price),
	          end_date = COALESCE($3, end_date)
	          WHERE id = $4`

	var endDate *time.Time
	if update.EndDate != nil {
		parsed, err := time.Parse("01-2006", *update.EndDate)
		if err != nil {
			r.logger.Error("Invalid end_date format",
				zap.String("end_date", *update.EndDate),
				zap.Error(err),
			)
			return fmt.Errorf("invalid end_date format: %v", err)
		}
		endDate = &parsed
	}

	_, err := r.db.ExecContext(ctx, query, update.ServiceName, update.Price, endDate, id)
	if err != nil {
		r.logger.Error("Failed to update subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
	}
	return err
}

// Delete удаляет подписку
// @Summary Удалить подписку
// @Description Удаляет подписку по указанному UUID
// @Param id path string true "UUID подписки"
// @Success 204
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("Deleting subscription", zap.String("id", id.String()))

	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete subscription",
			zap.String("id", id.String()),
			zap.Error(err),
		)
	}
	return err
}

// List возвращает список всех подписок
// @Summary Список подписок
// @Description Возвращает все подписки из базы данных
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) List(ctx context.Context) ([]*models.Subscription, error) {
	r.logger.Debug("Listing all subscriptions")

	var subs []*models.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`
	err := r.db.SelectContext(ctx, &subs, query)
	if err != nil {
		r.logger.Error("Failed to list subscriptions", zap.Error(err))
	}
	return subs, err
}

// CalculateTotalCost вычисляет суммарную стоимость подписок
// @Summary Расчет стоимости
// @Description Вычисляет суммарную стоимость подписок за указанный период с возможностью фильтрации
// @Param req body models.TotalCostRequest true "Параметры расчета"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse

func (r *SubscriptionRepo) CalculateTotalCost(ctx context.Context, req *models.TotalCostRequest) (int, error) {
	r.logger.Info("Calculating total cost",
		zap.String("from_date", req.FromDate),
		zap.String("to_date", req.ToDate),
		zap.Any("user_id", req.UserID),
		zap.Any("service_name", req.ServiceName),
	)

	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions 
	          WHERE start_date <= $1 AND (end_date IS NULL OR end_date >= $2)`

	args := []interface{}{parseDate(req.ToDate), parseDate(req.FromDate)}

	if req.UserID != nil {
		query += " AND user_id = $3"
		args = append(args, *req.UserID)
	}

	if req.ServiceName != nil {
		query += " AND service_name = $4"
		args = append(args, *req.ServiceName)
	}

	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		r.logger.Error("Failed to calculate total cost",
			zap.Error(err),
			zap.Any("request", req),
		)
	} else {
		r.logger.Debug("Total cost calculated",
			zap.Int("total", total),
			zap.Any("request", req),
		)
	}
	return total, err
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("01-2006", dateStr)
	return t
}
