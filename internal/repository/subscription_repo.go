package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type SubscriptionRepository struct {
	DB *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		DB: db,
	}
}

// Upsert creates or updates a subscription for a user.
func (r *SubscriptionRepository) UpsertSubscription(
	ctx context.Context,
	userID uuid.UUID,
	originalTransactionID string,
	productID string,
	status string,
	expiresAt time.Time,
	environment string,
) (*models.Subscription, error) {
	var sub models.Subscription

	err := r.DB.QueryRow(ctx, upsertSubscriptionQuery, userID, originalTransactionID, productID, status, expiresAt, environment).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.OriginalTransactionID,
		&sub.ProductID,
		&sub.Status,
		&sub.ExpiresAt,
		&sub.Environment,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to upsert subscription: %w", err)
	}

	return &sub, nil
}

// GetByUserID gets a subscription by user ID.
func (r *SubscriptionRepository) GetSubscriptionByUserID(
	ctx context.Context,
	userID uuid.UUID,
	viewerID uuid.UUID,
) (*models.Subscription, error) {
	var sub models.Subscription

	err := r.DB.QueryRow(ctx, getSubscriptionByUserIDQuery, userID, viewerID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.OriginalTransactionID,
		&sub.ProductID,
		&sub.Status,
		&sub.ExpiresAt,
		&sub.Environment,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// GetByTransactionID gets a subscription by original transaction ID.
func (r *SubscriptionRepository) GetSubscriptionByTransactionID(
	ctx context.Context,
	transactionID string,
	userID uuid.UUID,
) (*models.Subscription, error) {
	var sub models.Subscription

	err := r.DB.QueryRow(ctx, getSubscriptionByTransactionIDQuery, transactionID, userID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.OriginalTransactionID,
		&sub.ProductID,
		&sub.Status,
		&sub.ExpiresAt,
		&sub.Environment,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// Delete removes a subscription.
func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, userID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteSubscriptionByUserIDQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}
