package testutil

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

func InsertProfile(ctx context.Context, db *pgxpool.Pool, username string) (uuid.UUID, time.Time, error) {
	var id uuid.UUID
	var updatedAt time.Time
	err := db.QueryRow(
		ctx,
		"INSERT INTO profiles (username) VALUES ($1) RETURNING id, updated_at",
		username,
	).Scan(&id, &updatedAt)

	return id, updatedAt, err
}

func InsertSysAdmin(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) error {
	_, err := db.Exec(ctx, "INSERT INTO sys_admins (user_id) VALUES ($1)", userID)
	return err
}

func InsertMuscle(
	ctx context.Context,
	db *pgxpool.Pool,
	name string,
) (*models.Muscle, error) {
	var muscle models.Muscle
	err := db.QueryRow(
		ctx,
		"INSERT INTO muscles (name) VALUES ($1) RETURNING id, name, created_at",
		name,
	).Scan(&muscle.ID, &muscle.Name, &muscle.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &muscle, nil
}

// UpdateBuilder helps construct dynamic UPDATE queries with proper NULL handling.
type UpdateBuilder struct {
	sets  []string
	args  []interface{}
	index int
}

// NewUpdateBuilder creates a new UpdateBuilder starting at parameter index 1.
func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		sets:  make([]string, 0),
		args:  make([]interface{}, 0),
		index: 1,
	}
}

// AddField adds a field to the UPDATE query if the pointer is not nil.
// If the value is a zero value (empty string, zero time, etc.), it sets the DB column to NULL.
func (b *UpdateBuilder) AddField(column string, ptr interface{}) {
	if ptr == nil || reflect.ValueOf(ptr).IsNil() {
		return
	}

	val := reflect.ValueOf(ptr).Elem()
	b.sets = append(b.sets, fmt.Sprintf("%s = $%d", column, b.index))

	// Check for zero values that should be converted to NULL
	switch v := val.Interface().(type) {
	case string:
		if v == "" {
			b.args = append(b.args, nil)
		} else {
			b.args = append(b.args, v)
		}
	case time.Time:
		if v.IsZero() {
			b.args = append(b.args, nil)
		} else {
			b.args = append(b.args, v)
		}
	default:
		b.args = append(b.args, val.Interface())
	}
	b.index++
}

// AddFieldNoNullConvert adds a field without converting zero values to NULL.
// Useful for boolean fields where false is a valid value.
func (b *UpdateBuilder) AddFieldNoNullConvert(column string, ptr interface{}) {
	if ptr == nil || reflect.ValueOf(ptr).IsNil() {
		return
	}

	val := reflect.ValueOf(ptr).Elem()
	b.sets = append(b.sets, fmt.Sprintf("%s = $%d", column, b.index))
	b.args = append(b.args, val.Interface())
	b.index++
}

// HasUpdates returns true if there are any fields to update.
func (b *UpdateBuilder) HasUpdates() bool {
	return len(b.sets) > 0
}

// Build constructs the SET clause and returns it along with the args.
// The caller should append their WHERE clause parameters to args.
func (b *UpdateBuilder) Build() (setClause string, args []interface{}, nextIndex int) {
	return strings.Join(b.sets, ", "), b.args, b.index
}

// BuildQuery constructs the full UPDATE query.
func (b *UpdateBuilder) BuildQuery(table string, whereColumn string, whereValue interface{}) (query string, args []interface{}) {
	setClause, queryArgs, nextIndex := b.Build()
	query = fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", table, setClause, whereColumn, nextIndex)
	queryArgs = append(queryArgs, whereValue)
	return query, queryArgs
}
