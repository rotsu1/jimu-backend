package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *UserRepository) GetProfileByID(
	ctx context.Context,
	viewerID uuid.UUID,
	targetID uuid.UUID,
) (*models.Profile, error) {
	var profile models.Profile

	err := r.DB.QueryRow(ctx, getProfileByIDQuery, viewerID, targetID).Scan(
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
		&profile.FollowersCount,
		&profile.FollowingCount,
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
	id uuid.UUID,
	updates models.UpdateProfileRequest,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.Username != nil {
		sets = append(sets, fmt.Sprintf("username = $%d", i))

		if *updates.Username == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Username)
		}
		i++
	}
	if updates.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name = $%d", i))
		if *updates.DisplayName == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.DisplayName)
		}
		i++
	}
	if updates.Bio != nil {
		sets = append(sets, fmt.Sprintf("bio = $%d", i))
		if *updates.Bio == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Bio)
		}
		i++
	}
	if updates.Location != nil {
		sets = append(sets, fmt.Sprintf("location = $%d", i))
		if *updates.Location == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Location)
		}
		i++
	}
	if updates.BirthDate != nil {
		sets = append(sets, fmt.Sprintf("birth_date = $%d", i))
		if (*updates.BirthDate).IsZero() {
			args = append(args, nil)
		} else {
			args = append(args, *updates.BirthDate)
		}
		i++
	}
	if updates.AvatarURL != nil {
		sets = append(sets, fmt.Sprintf("avatar_url = $%d", i))
		if *updates.AvatarURL == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.AvatarURL)
		}
		i++
	}
	if updates.SubscriptionPlan != nil {
		sets = append(sets, fmt.Sprintf("subscription_plan = $%d", i))
		if *updates.SubscriptionPlan == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.SubscriptionPlan)
		}
		i++
	}
	if updates.IsPrivateAccount != nil {
		sets = append(sets, fmt.Sprintf("is_private_account = $%d", i))
		args = append(args, *updates.IsPrivateAccount)
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE profiles SET %s WHERE id = $%d",
		strings.Join(sets, ", "),
		i,
	)
	args = append(args, id)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUsernameTaken
		}
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *UserRepository) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteProfileByIDQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *UserRepository) GetUserSettingsByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.UserSetting, error) {

	var userSetting models.UserSetting

	err := r.DB.QueryRow(ctx, getUserSettingsByIDQuery, id).Scan(
		&userSetting.UserID,
		&userSetting.NotifyNewFollower,
		&userSetting.NotifyLikes,
		&userSetting.NotifyComments,
		&userSetting.SoundEnabled,
		&userSetting.SoundEffectName,
		&userSetting.DefaultTimerSeconds,
		&userSetting.AutoFillPreviousValues,
		&userSetting.UnitWeight,
		&userSetting.UnitDistance,
		&userSetting.UnitLength,
		&userSetting.CreatedAt,
		&userSetting.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &userSetting, nil
}

func (r *UserRepository) UpdateUserSettings(
	ctx context.Context,
	id string,
	updates models.UpdateUserSettingsRequest,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.NotifyNewFollower != nil {
		sets = append(sets, fmt.Sprintf("notify_new_follower = $%d", i))
		args = append(args, *updates.NotifyNewFollower)
		i++
	}
	if updates.NotifyLikes != nil {
		sets = append(sets, fmt.Sprintf("notify_likes = $%d", i))
		args = append(args, *updates.NotifyLikes)
		i++
	}
	if updates.NotifyComments != nil {
		sets = append(sets, fmt.Sprintf("notify_comments = $%d", i))
		args = append(args, *updates.NotifyComments)
		i++
	}
	if updates.SoundEnabled != nil {
		sets = append(sets, fmt.Sprintf("sound_enabled = $%d", i))
		args = append(args, *updates.SoundEnabled)
		i++
	}
	if updates.SoundEffectName != nil {
		sets = append(sets, fmt.Sprintf("sound_effect_name = $%d", i))
		args = append(args, *updates.SoundEffectName)
		i++
	}
	if updates.DefaultTimerSeconds != nil {
		sets = append(sets, fmt.Sprintf("default_timer_seconds = $%d", i))
		args = append(args, *updates.DefaultTimerSeconds)
		i++
	}
	if updates.AutoFillPreviousValues != nil {
		sets = append(sets, fmt.Sprintf("auto_fill_previous_values = $%d", i))
		args = append(args, *updates.AutoFillPreviousValues)
		i++
	}
	if updates.UnitWeight != nil {
		sets = append(sets, fmt.Sprintf("unit_weight = $%d", i))
		args = append(args, *updates.UnitWeight)
		i++
	}
	if updates.UnitDistance != nil {
		sets = append(sets, fmt.Sprintf("unit_distance = $%d", i))
		args = append(args, *updates.UnitDistance)
		i++
	}
	if updates.UnitLength != nil {
		sets = append(sets, fmt.Sprintf("unit_length = $%d", i))
		args = append(args, *updates.UnitLength)
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE user_settings SET %s WHERE user_id = $%d",
		strings.Join(sets, ", "),
		i,
	)
	args = append(args, id)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *UserRepository) GetIdentitiesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*models.UserIdentity, error) {
	rows, err := r.DB.Query(ctx, getIdentitiesByUserIDQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identities: %w", err)
	}
	defer rows.Close()

	var identities []*models.UserIdentity
	for rows.Next() {
		var i models.UserIdentity
		err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ProviderName,
			&i.ProviderUserID,
			&i.ProviderEmail,
			&i.LastSignInAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan identity: %w", err)
		}
		identities = append(identities, &i)
	}
	return identities, nil
}

func (r *UserRepository) DeleteIdentity(
	ctx context.Context,
	userID uuid.UUID,
	provider string,
) error {
	res, err := r.DB.Exec(ctx, deleteIdentityQuery, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
