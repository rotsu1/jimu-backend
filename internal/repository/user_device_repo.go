package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type UserDeviceRepository struct {
	DB *pgxpool.Pool
}

func NewUserDeviceRepository(db *pgxpool.Pool) *UserDeviceRepository {
	return &UserDeviceRepository{
		DB: db,
	}
}

// Upsert creates or updates a user device by FCM token.
func (r *UserDeviceRepository) UpsertUserDevice(
	ctx context.Context,
	userID uuid.UUID,
	fcmToken string,
	deviceType string,
	viewerID uuid.UUID,
) (*models.UserDevice, error) {
	var device models.UserDevice

	err := r.DB.QueryRow(ctx, upsertUserDeviceQuery, userID, fcmToken, deviceType, viewerID).Scan(
		&device.ID,
		&device.UserID,
		&device.FCMToken,
		&device.DeviceType,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to upsert user device: %w", err)
	}

	return &device, nil
}

// GetByID gets a device by ID.
func (r *UserDeviceRepository) GetUserDeviceByID(
	ctx context.Context,
	deviceID uuid.UUID,
	userID uuid.UUID,
) (*models.UserDevice, error) {
	var device models.UserDevice

	err := r.DB.QueryRow(ctx, getUserDeviceByIDQuery, deviceID, userID).Scan(
		&device.ID,
		&device.UserID,
		&device.FCMToken,
		&device.DeviceType,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserDeviceNotFound
		}
		return nil, fmt.Errorf("failed to get user device: %w", err)
	}

	return &device, nil
}

// GetByUserID gets all devices for a user.
func (r *UserDeviceRepository) GetUserDevicesByUserID(
	ctx context.Context,
	targetID uuid.UUID,
	viewerID uuid.UUID,
) ([]*models.UserDevice, error) {
	rows, err := r.DB.Query(ctx, getUserDevicesByUserIDQuery, targetID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.UserDevice
	for rows.Next() {
		var d models.UserDevice
		err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.FCMToken,
			&d.DeviceType,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user device: %w", err)
		}
		devices = append(devices, &d)
	}

	return devices, nil
}

// GetByFCMToken gets a device by FCM token.
func (r *UserDeviceRepository) GetUserDeviceByFCMToken(
	ctx context.Context,
	fcmToken string,
	userID uuid.UUID,
) (*models.UserDevice, error) {
	var device models.UserDevice

	err := r.DB.QueryRow(ctx, getUserDeviceByFCMTokenQuery, fcmToken, userID).Scan(
		&device.ID,
		&device.UserID,
		&device.FCMToken,
		&device.DeviceType,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserDeviceNotFound
		}
		return nil, fmt.Errorf("failed to get user device: %w", err)
	}

	return &device, nil
}

// Delete removes a device by ID.
func (r *UserDeviceRepository) DeleteUserDeviceByID(
	ctx context.Context,
	deviceID uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteUserDeviceByIDQuery, deviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user device: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserDeviceNotFound
	}

	return nil
}

// DeleteByUserID removes all devices for a user.
func (r *UserDeviceRepository) DeleteUserDevicesByUserID(
	ctx context.Context,
	targetID uuid.UUID,
	viewerID uuid.UUID,
) error {
	_, err := r.DB.Exec(ctx, deleteUserDevicesByUserIDQuery, targetID, viewerID)
	if err != nil {
		return fmt.Errorf("failed to delete user devices: %w", err)
	}
	return nil
}
