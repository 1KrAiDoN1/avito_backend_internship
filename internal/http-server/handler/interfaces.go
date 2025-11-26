package handler

import (
	"context"
	"internship/internal/domain/entity"
)

type PullRequestServiceInterface interface {
	CreatePullRequest(ctx context.Context, pr *entity.PullRequest) (*entity.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*entity.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entity.PullRequest, string, error)
}

type UserServiceInterface interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	GetReviewPullRequests(ctx context.Context, userID string) ([]entity.PullRequestShort, error)
	DeactivateTeamMembers(ctx context.Context, teamName string) ([]entity.PullRequest, error)
}

type TeamServiceInterface interface {
	CreateTeam(ctx context.Context, team *entity.Team) (*entity.Team, error)
	IsTeamExists(ctx context.Context, teamName string) (bool, error)
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
}

type StatisticsServiceInterface interface {
	GetAssignmentStats(ctx context.Context) (map[string]int, error)
	GetPRStats(ctx context.Context) (map[string]interface{}, error)
	GetFullStats(ctx context.Context) (map[string]interface{}, error)
}
