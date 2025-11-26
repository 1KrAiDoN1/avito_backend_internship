package service

import (
	"context"
	"internship/internal/domain/entity"
)

type TeamRepositoryInterface interface {
	Create(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, teamName string) (*entity.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

// UserRepository определяет интерфейс для работы с пользователями
type UserRepositoryInterface interface {
	BatchCreateOrUpdate(ctx context.Context, users []*entity.User) error
	Update(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID string) (*entity.User, error)
	GetByTeamName(ctx context.Context, teamName string) ([]entity.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	DeactivateTeamMembers(ctx context.Context, teamName string) error
}

// PullRequestRepository определяет интерфейс для работы с PR
type PullRequestRepositoryInterface interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	GetByID(ctx context.Context, prID string) (*entity.PullRequest, error)
	Update(ctx context.Context, pr *entity.PullRequest) error
	Exists(ctx context.Context, prID string) (bool, error)
	GetByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error)
	GetOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]entity.PullRequest, error)
}

// ReviewerRepository определяет интерфейс для работы с назначениями ревьюверов
type ReviewerRepositoryInterface interface {
	AssignReviewer(ctx context.Context, prID, userID string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	IsAssigned(ctx context.Context, prID, userID string) (bool, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
}

// StatisticsRepository определяет интерфейс для получения статистики
type StatisticsRepositoryInterface interface {
	GetAssignmentStats(ctx context.Context) (map[string]int, error)
	GetPRStats(ctx context.Context) (map[string]interface{}, error)
}
