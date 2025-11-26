package postgres

import (
	"context"
	"errors"
	"fmt"
	"internship/internal/domain/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryCreatePR = `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	queryGetPRByID = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	queryUpdatePR = `
		UPDATE pull_requests
		SET pull_request_name = $2, status = $3, merged_at = $4
		WHERE pull_request_id = $1
	`

	queryExistsPR = `
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`

	queryGetByReviewer = `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	queryGetOpenPRsByReviewers = `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = ANY($1) AND pr.status = $2
		ORDER BY pr.created_at DESC
	`
)

type PullRequestRepository struct {
	pool         *pgxpool.Pool
	reviewerRepo *ReviewerRepository
}

func NewPullRequestRepository(pool *pgxpool.Pool, reviewerRepo *ReviewerRepository) *PullRequestRepository {
	return &PullRequestRepository{
		pool:         pool,
		reviewerRepo: reviewerRepo,
	}
}

// Create создает новый PR
func (r *PullRequestRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, queryCreatePR,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		now,
		pr.MergedAt,
	)

	if err != nil {
		return fmt.Errorf("create pull request: %w", err)
	}

	return nil
}

// GetByID получает PR по ID с ревьюверами
func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (*entity.PullRequest, error) {
	var pr entity.PullRequest
	err := r.pool.QueryRow(ctx, queryGetPRByID, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrPRNotFound
		}
		return nil, fmt.Errorf("get pull request by id: %w", err)
	}

	// Получаем ревьюверов
	reviewers, err := r.reviewerRepo.GetReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

// Update обновляет PR
func (r *PullRequestRepository) Update(ctx context.Context, pr *entity.PullRequest) error {
	result, err := r.pool.Exec(ctx, queryUpdatePR,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.Status,
		pr.MergedAt,
	)

	if err != nil {
		return fmt.Errorf("update pull request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrPRNotFound
	}

	return nil
}

// Exists проверяет существование PR
func (r *PullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, queryExistsPR, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check pr exists: %w", err)
	}

	return exists, nil
}

// GetByReviewer получает PR'ы где пользователь назначен ревьювером
func (r *PullRequestRepository) GetByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error) {
	rows, err := r.pool.Query(ctx, queryGetByReviewer, userID)
	if err != nil {
		return nil, fmt.Errorf("get prs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		// Получаем ревьюверов для каждого PR
		reviewers, err := r.reviewerRepo.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers for pr %s: %w", pr.PullRequestID, err)
		}
		pr.AssignedReviewers = reviewers

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull requests: %w", err)
	}

	return prs, nil
}

// GetOpenPRsByReviewers получает открытые PR'ы для списка ревьюверов
func (r *PullRequestRepository) GetOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]entity.PullRequest, error) {
	if len(reviewerIDs) == 0 {
		return []entity.PullRequest{}, nil
	}

	rows, err := r.pool.Query(ctx, queryGetOpenPRsByReviewers, reviewerIDs, entity.PRStatusOpen)
	if err != nil {
		return nil, fmt.Errorf("get open prs by reviewers: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		// Получаем ревьюверов
		reviewers, err := r.reviewerRepo.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers: %w", err)
		}
		pr.AssignedReviewers = reviewers

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull requests: %w", err)
	}

	return prs, nil
}
