package repository

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type UserSessionRepository struct {
	DB *pgxpool.Pool
}

func NewUserSessionRepository(db *pgxpool.Pool) *UserSessionRepository {
	return &UserSessionRepository{
		DB: db,
	}
}

// Create creates a new session.
func (r *UserSessionRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	userAgent *string,
	clientIP *netip.Addr,
	expiresAt time.Time,
) (*models.UserSession, error) {
	var session models.UserSession

	err := r.DB.QueryRow(ctx, insertUserSessionQuery, userID, refreshToken, userAgent, clientIP, expiresAt).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIP,
		&session.IsRevoked,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			case "23505":
				return nil, ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetByID gets a session by ID with access control.
func (r *UserSessionRepository) GetSessionByID(
	ctx context.Context,
	id uuid.UUID,
	viewerID uuid.UUID,
) (*models.UserSession, error) {
	var session models.UserSession

	// Pass viewerID for guard
	err := r.DB.QueryRow(ctx, getUserSessionByIDQuery, id, viewerID).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIP,
		&session.IsRevoked,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetByRefreshToken gets an active (non-revoked, non-expired) session by refresh token.
func (r *UserSessionRepository) GetSessionByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*models.UserSession, error) {
	var session models.UserSession

	err := r.DB.QueryRow(ctx, getUserSessionByRefreshTokenQuery, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIP,
		&session.IsRevoked,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetByUserID gets all sessions for a user with access control (Owner or Admin).
func (r *UserSessionRepository) GetSessionsByUserID(
	ctx context.Context,
	targetUserID uuid.UUID,
	viewerID uuid.UUID,
) ([]*models.UserSession, error) {
	// guarded query takes target ($1) and viewer ($2)
	rows, err := r.DB.Query(ctx, getUserSessionsByUserIDQuery, targetUserID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.UserSession
	for rows.Next() {
		var s models.UserSession
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.RefreshToken,
			&s.UserAgent,
			&s.ClientIP,
			&s.IsRevoked,
			&s.ExpiresAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &s)
	}

	return sessions, nil
}

// Revoke revokes a specific session.
func (r *UserSessionRepository) RevokeSession(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, revokeUserSessionQuery, id, viewerID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserSessionNotFound
	}

	return nil
}

// RevokeAllForUser revokes all sessions for a user.
func (r *UserSessionRepository) RevokeAllSessionsForUser(ctx context.Context, targetUserID uuid.UUID, viewerID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, revokeUserSessionsByUserIDQuery, targetUserID, viewerID)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return nil
}

// DeleteExpired deletes all expired sessions.
func (r *UserSessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	commandTag, err := r.DB.Exec(ctx, deleteExpiredUserSessionsQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return commandTag.RowsAffected(), nil
}

// Delete removes a session by ID.
func (r *UserSessionRepository) DeleteSession(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteUserSessionByIDQuery, id, viewerID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserSessionNotFound
	}

	return nil
}
