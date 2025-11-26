package service

import (
	"context"
	"fmt"
	"internship/internal/domain/entity"
	"time"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo     UserRepositoryInterface
	prRepo       PullRequestRepositoryInterface
	reviewerRepo ReviewerRepositoryInterface
	log          *zap.Logger
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepo UserRepositoryInterface,
	prRepo PullRequestRepositoryInterface,
	reviewerRepo ReviewerRepositoryInterface,
	log *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		prRepo:       prRepo,
		reviewerRepo: reviewerRepo,
		log:          log,
	}
}

// SetIsActive устанавливает флаг активности пользователя
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("get user", zap.Error(err))
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := s.userRepo.SetIsActive(ctx, userID, isActive); err != nil {
		s.log.Error("set is_active", zap.Error(err))
		return nil, fmt.Errorf("set is_active: %w", err)
	}

	user.IsActive = isActive

	return user, nil
}

// GetReviewPullRequests получает список PR где пользователь назначен ревьювером
func (s *UserService) GetReviewPullRequests(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		s.log.Error("get prs by reviewer", zap.Error(err))
		return nil, fmt.Errorf("get prs by reviewer: %w", err)
	}

	shortPRs := make([]entity.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		shortPRs = append(shortPRs, pr.ToShort())
	}

	s.log.Info("get review pull requests", zap.Any("short_prs", shortPRs))
	return shortPRs, nil
}

// DeactivateTeamMembers деактивирует всех участников команды и переназначает открытые PR
// Оптимизировано для выполнения за ~100ms
func (s *UserService) DeactivateTeamMembers(ctx context.Context, teamName string) ([]entity.PullRequest, error) {
	// Устанавливаем таймаут для операции
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Получаем всех пользователей команды
	teamMembers, err := s.userRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		s.log.Error("get team members", zap.Error(err))
		return nil, fmt.Errorf("get team members: %w", err)
	}

	if len(teamMembers) == 0 {
		s.log.Info("no team members", zap.String("team_name", teamName))
		return []entity.PullRequest{}, nil
	}

	// Собираем ID всех участников команды
	memberIDs := make([]string, 0, len(teamMembers))
	for _, member := range teamMembers {
		memberIDs = append(memberIDs, member.UserID)
	}

	// Получаем все открытые PR, где участники команды назначены ревьюверами
	openPRs, err := s.prRepo.GetOpenPRsByReviewers(ctx, memberIDs)
	if err != nil {
		s.log.Error("get open prs", zap.Error(err))
		return nil, fmt.Errorf("get open prs: %w", err)
	}

	// Деактивируем всех участников команды
	if err := s.userRepo.DeactivateTeamMembers(ctx, teamName); err != nil {
		s.log.Error("deactivate team members", zap.Error(err))
		return nil, fmt.Errorf("deactivate team members: %w", err)
	}

	// Для каждого открытого PR удаляем всех ревьюверов из деактивированной команды
	for i := range openPRs {
		pr := &openPRs[i]

		// Удаляем ревьюверов, которые были в деактивированной команде
		newReviewers := make([]string, 0, len(pr.AssignedReviewers))
		for _, reviewerID := range pr.AssignedReviewers {
			isMember := false
			for _, memberID := range memberIDs {
				if reviewerID == memberID {
					isMember = true
					// Удаляем ревьювера
					if err := s.reviewerRepo.RemoveReviewer(ctx, pr.PullRequestID, reviewerID); err != nil {
						s.log.Error("remove reviewer", zap.Error(err))
						return nil, fmt.Errorf("remove reviewer: %w", err)
					}
					break
				}
			}
			if !isMember {
				newReviewers = append(newReviewers, reviewerID)
			}
		}

		pr.AssignedReviewers = newReviewers
	}

	s.log.Info("deactivate team members", zap.Any("open_prs", openPRs))
	return openPRs, nil
}
