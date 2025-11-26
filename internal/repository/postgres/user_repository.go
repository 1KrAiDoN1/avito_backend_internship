package postgres

import (
	"context"
	"errors"
	"fmt"
	"internship/internal/domain/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryCreateOrUpdateUser = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	querySetIsActive = `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
	`

	queryUpdate = `
		UPDATE users
		SET username = $2, team_name = $3, is_active = $4
		WHERE user_id = $1
	`

	queryDeactivateTeamMembers = `
		UPDATE users
		SET is_active = false
		WHERE team_name = $1
	`

	queryGetByID = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	queryGetByTeamName = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// BatchCreateOrUpdate создает или обновляет пользователей
func (r *UserRepository) BatchCreateOrUpdate(ctx context.Context, users []*entity.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, user := range users {
		_, err := tx.Exec(ctx, queryCreateOrUpdateUser,
			user.UserID,
			user.Username,
			user.TeamName,
			user.IsActive,
		)
		if err != nil {
			return fmt.Errorf("update user %s: %w", user.UserID, err)
		}

	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {

	result, err := r.pool.Exec(ctx, queryUpdate,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	var user entity.User
	err := r.pool.QueryRow(ctx, queryGetByID, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// GetByTeamName получает всех пользователей команды
func (r *UserRepository) GetByTeamName(ctx context.Context, teamName string) ([]entity.User, error) {

	rows, err := r.pool.Query(ctx, queryGetByTeamName, teamName)
	if err != nil {
		return nil, fmt.Errorf("get users by team: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

// SetIsActive устанавливает флаг активности пользователя
func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {

	result, err := r.pool.Exec(ctx, querySetIsActive, userID, isActive)
	if err != nil {
		return fmt.Errorf("set is_active: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// DeactivateTeamMembers деактивирует всех участников команды
func (r *UserRepository) DeactivateTeamMembers(ctx context.Context, teamName string) error {

	_, err := r.pool.Exec(ctx, queryDeactivateTeamMembers, teamName)
	if err != nil {
		return fmt.Errorf("deactivate team members: %w", err)
	}

	return nil
}
