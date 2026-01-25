package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) UpsertGoogleUser(
	ctx context.Context,
	googleID string,
	email string,
) (*models.Profile, error) {
	// Check if the Google account is already registered in the Identity table
	identity, err := r.GetIdentityByProvider(ctx, "google", googleID)

	if err == pgx.ErrNoRows {
		// If the user is new, insert the record into the Identity table
		// This will trigger the DB trigger that creates the Profile automatically
		var userID uuid.UUID
		err = r.DB.QueryRow(ctx, insertUserIdentityQuery, googleID, email).Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create new identity: %w", err)
		}
		return &models.Profile{ID: userID}, nil
	} else if err != nil {
		return nil, err
	} else {
		// If the user is existing, update the last sign in time
		_, err = r.DB.Exec(ctx, updateUserIdentityQuery, googleID, email)
		if err != nil {
			return nil, fmt.Errorf("failed to update login time: %w", err)
		}
	}

	// Return the Profile with the userID
	return &models.Profile{ID: identity.UserID}, nil
}

func (r *UserRepository) GetIdentityByProvider(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*models.UserIdentity, error) {
	var userIdentity models.UserIdentity

	err := r.DB.QueryRow(ctx, getUserIdentityByProviderQuery, provider, providerUserID).Scan(
		&userIdentity.ID,
		&userIdentity.UserID,
		&userIdentity.ProviderName,
		&userIdentity.ProviderUserID,
		&userIdentity.ProviderEmail,
		&userIdentity.LastSignInAt,
		&userIdentity.CreatedAt,
		&userIdentity.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &userIdentity, nil
}

func (r *UserRepository) GetProfileByID(ctx context.Context, id string) (*models.Profile, error) {
	var profile models.Profile

	err := r.DB.QueryRow(ctx, getProfileQuery, id).Scan(
		&profile.ID,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.Location,
		&profile.BirthDate,
		&profile.SubscriptionPlan,
		&profile.IsPrivateAccount,
		&profile.LastWorkedOutAt,
		&profile.TotalWorkouts,
		&profile.CurrentStreak,
		&profile.TotalWeight,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *UserRepository) UpdateProfile(
	ctx context.Context,
	id string,
	updates models.UpdateProfileRequest,
) error {
	var sets []string
	var args []interface{}
	i := 1

	// 各フィールドのチェック
	if updates.Username != nil {
		sets = append(sets, fmt.Sprintf("username = $%d", i))
		args = append(args, *updates.Username) // ポインタの中身を渡す
		i++
	}
	if updates.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name = $%d", i))
		args = append(args, *updates.DisplayName)
		i++
	}
	if updates.Bio != nil {
		sets = append(sets, fmt.Sprintf("bio = $%d", i))
		args = append(args, *updates.Bio)
		i++
	}
	if updates.Location != nil {
		sets = append(sets, fmt.Sprintf("location = $%d", i))
		args = append(args, *updates.Location)
		i++
	}
	if updates.BirthDate != nil {
		sets = append(sets, fmt.Sprintf("birth_date = $%d", i))
		args = append(args, *updates.BirthDate)
		i++
	}
	if updates.AvatarURL != nil {
		sets = append(sets, fmt.Sprintf("avatar_url = $%d", i))
		args = append(args, *updates.AvatarURL)
		i++
	}
	if updates.SubscriptionPlan != nil {
		sets = append(sets, fmt.Sprintf("subscription_plan = $%d", i))
		args = append(args, *updates.SubscriptionPlan)
		i++
	}
	if updates.IsPrivateAccount != nil {
		sets = append(sets, fmt.Sprintf("is_private_account = $%d", i))
		args = append(args, *updates.IsPrivateAccount)
		i++
	}

	query := fmt.Sprintf(
		"UPDATE profiles SET %s WHERE id = $%d",
		strings.Join(sets, ", "),
		i,
	)
	args = append(args, id)

	_, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}
