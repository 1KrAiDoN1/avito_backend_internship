package postgres

import (
	"context"
	"errors"
	"fmt"
	"internship/domain/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryCreateTeam      = `INSERT INTO teams (team_name) VALUES ($1)`
	queryGetTeamByName   = `SELECT team_name FROM teams WHERE team_name = $1`
	queryCheckTeamExists = `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

// Create создает новую команду
func (r *TeamRepository) Create(ctx context.Context, team *entity.Team) error {

	_, err := r.pool.Exec(ctx, queryCreateTeam, team.TeamName)
	if err != nil {
		return fmt.Errorf("create team: %w", err)
	}

	return nil
}

// GetByName получает команду по имени
func (r *TeamRepository) GetByName(ctx context.Context, teamName string) (*entity.Team, error) {
	var team entity.Team
	err := r.pool.QueryRow(ctx, queryGetTeamByName, teamName).Scan(
		&team.TeamName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrTeamNotFound
		}
		return nil, fmt.Errorf("get team by name: %w", err)
	}

	return &team, nil
}

// Exists проверяет существование команды
func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, queryCheckTeamExists, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check team exists: %w", err)
	}

	return exists, nil
}
