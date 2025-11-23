package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryAssignReviewer = `
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (pull_request_id, user_id) DO NOTHING
	`
	queryRemoveReviewer = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND user_id = $2
	`
	queryGetReviewers = `
		SELECT user_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at
	`
	queryIsAssigned = `
		SELECT EXISTS(SELECT 1 FROM pull_request_reviewers WHERE pull_request_id = $1 AND user_id = $2)
	`
	queryReplaceReviewer = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND user_id = $2
	`
	queryInsertReviewer = `
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (pull_request_id, user_id) DO NOTHING
	`
	queryDeleteReviewer = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND user_id = $2
	`
)

type ReviewerRepository struct {
	pool *pgxpool.Pool
}

func NewReviewerRepository(pool *pgxpool.Pool) *ReviewerRepository {
	return &ReviewerRepository{pool: pool}
}

// AssignReviewer назначает ревьювера на PR
func (r *ReviewerRepository) AssignReviewer(ctx context.Context, prID, userID string) error {
	_, err := r.pool.Exec(ctx, queryAssignReviewer, prID, userID)
	if err != nil {
		return fmt.Errorf("assign reviewer: %w", err)
	}

	return nil
}

// RemoveReviewer удаляет ревьювера с PR
func (r *ReviewerRepository) RemoveReviewer(ctx context.Context, prID, userID string) error {
	_, err := r.pool.Exec(ctx, queryRemoveReviewer, prID, userID)
	if err != nil {
		return fmt.Errorf("remove reviewer: %w", err)
	}

	return nil
}

// GetReviewers получает список ревьюверов для PR
func (r *ReviewerRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, queryGetReviewers, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	return reviewers, nil
}

// IsAssigned проверяет, назначен ли пользователь ревьювером на PR
func (r *ReviewerRepository) IsAssigned(ctx context.Context, prID, userID string) (bool, error) {
	var assigned bool
	err := r.pool.QueryRow(ctx, queryIsAssigned, prID, userID).Scan(&assigned)
	if err != nil {
		return false, fmt.Errorf("check is assigned: %w", err)
	}

	return assigned, nil
}

// ReplaceReviewer заменяет одного ревьювера на другого в транзакции
func (r *ReviewerRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Удаляем старого ревьювера
	_, err = tx.Exec(ctx, queryDeleteReviewer, prID, oldUserID)
	if err != nil {
		return fmt.Errorf("remove old reviewer: %w", err)
	}

	// Добавляем нового ревьювера
	_, err = tx.Exec(ctx, queryInsertReviewer, prID, newUserID)
	if err != nil {
		return fmt.Errorf("assign new reviewer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
