package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryGetAssignmentStats = `
		SELECT u.username, COUNT(prr.pull_request_id) as assignment_count
		FROM users u
		LEFT JOIN pull_request_reviewers prr ON u.user_id = prr.user_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC
	`
	queryGetPRStats = `
		SELECT
			COUNT(*) as total_prs,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) as open_prs,
			COUNT(CASE WHEN status = 'MERGED' THEN 1 END) as merged_prs
		FROM pull_requests
	`
)

type StatisticsRepository struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{pool: pool}
}

// GetAssignmentStats возвращает статистику назначений по пользователям
func (r *StatisticsRepository) GetAssignmentStats(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, queryGetAssignmentStats)
	if err != nil {
		return nil, fmt.Errorf("get assignment stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var username string
		var count int
		if err := rows.Scan(&username, &count); err != nil {
			return nil, fmt.Errorf("scan assignment stat: %w", err)
		}
		stats[username] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assignment stats: %w", err)
	}

	return stats, nil
}

// GetPRStats возвращает общую статистику по PR
func (r *StatisticsRepository) GetPRStats(ctx context.Context) (map[string]interface{}, error) {
	var totalPRs, openPRs, mergedPRs int
	err := r.pool.QueryRow(ctx, queryGetPRStats).Scan(&totalPRs, &openPRs, &mergedPRs)
	if err != nil {
		return nil, fmt.Errorf("get pr stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_prs":  totalPRs,
		"open_prs":   openPRs,
		"merged_prs": mergedPRs,
	}

	return stats, nil
}
