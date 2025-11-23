package service

import (
	"context"
	"fmt"
	"internship/domain/entity"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type PullRequestService struct {
	prRepo       PullRequestRepositoryInterface
	userRepo     UserRepositoryInterface
	reviewerRepo ReviewerRepositoryInterface
	log          *zap.Logger
}

func NewPullRequestService(
	prRepo PullRequestRepositoryInterface,
	userRepo UserRepositoryInterface,
	reviewerRepo ReviewerRepositoryInterface,
	log *zap.Logger,
) *PullRequestService {
	return &PullRequestService{
		prRepo:       prRepo,
		userRepo:     userRepo,
		reviewerRepo: reviewerRepo,
		log:          log,
	}
}

// CreatePullRequest создает PR и автоматически назначает до 2 ревьюверов
func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr *entity.PullRequest) (*entity.PullRequest, error) {

	exists, err := s.prRepo.Exists(ctx, pr.PullRequestID)
	if err != nil {
		s.log.Error("check pr exists", zap.Error(err))
		return nil, fmt.Errorf("check pr exists: %w", err)
	}

	if exists {
		s.log.Error("pr exists", zap.String("pr_id", pr.PullRequestID))
		return nil, entity.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		s.log.Error("get author", zap.Error(err))
		return nil, fmt.Errorf("get author: %w", err)
	}

	teamMembers, err := s.userRepo.GetByTeamName(ctx, author.TeamName)
	if err != nil {
		s.log.Error("get team members", zap.Error(err))
		return nil, fmt.Errorf("get team members: %w", err)
	}

	candidates := make([]entity.User, 0)
	for _, member := range teamMembers {
		if member.UserID != pr.AuthorID && member.IsActive {
			candidates = append(candidates, member)
		}
	}

	reviewers := s.selectRandomReviewers(candidates, 2)

	pr.Status = entity.PRStatusOpen
	now := time.Now()
	pr.CreatedAt = &now

	if err := s.prRepo.Create(ctx, pr); err != nil {
		s.log.Error("create pr", zap.Error(err))
		return nil, fmt.Errorf("create pr: %w", err)
	}

	for _, reviewer := range reviewers {
		if err := s.reviewerRepo.AssignReviewer(ctx, pr.PullRequestID, reviewer.UserID); err != nil {
			s.log.Error("assign reviewer", zap.Error(err))
			return nil, fmt.Errorf("assign reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewer.UserID)
	}

	return pr, nil
}

// MergePullRequest помечает PR как MERGED (идемпотентная операция)
func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*entity.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("get pr", zap.Error(err))
		return nil, fmt.Errorf("get pr: %w", err)
	}

	if pr.Status == entity.PRStatusMerged {
		return pr, nil
	}
	pr.Status = entity.PRStatusMerged
	mergedAt := time.Now()
	pr.MergedAt = &mergedAt

	if err := s.prRepo.Update(ctx, pr); err != nil {
		s.log.Error("update pr", zap.Error(err))
		return nil, fmt.Errorf("update pr: %w", err)
	}

	return pr, nil
}

// ReassignReviewer переназначает ревьювера на другого из его команды
func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entity.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("get pr", zap.Error(err))
		return nil, "", fmt.Errorf("get pr: %w", err)
	}

	// Проверяем, что PR не в статусе MERGED
	if pr.Status == entity.PRStatusMerged {
		s.log.Error("pr merged", zap.String("pr_id", pr.PullRequestID))
		return nil, "", entity.ErrPRMerged
	}

	isAssigned, err := s.reviewerRepo.IsAssigned(ctx, prID, oldUserID)
	if err != nil {
		s.log.Error("check is assigned", zap.Error(err))
		return nil, "", fmt.Errorf("check is assigned: %w", err)
	}

	if !isAssigned {
		s.log.Error("not assigned", zap.String("pr_id", pr.PullRequestID), zap.String("old_user_id", oldUserID))
		return nil, "", entity.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		s.log.Error("get old reviewer", zap.Error(err))
		return nil, "", fmt.Errorf("get old reviewer: %w", err)
	}

	teamMembers, err := s.userRepo.GetByTeamName(ctx, oldReviewer.TeamName)
	if err != nil {
		s.log.Error("get team members", zap.Error(err))
		return nil, "", fmt.Errorf("get team members: %w", err)
	}

	// Фильтруем кандидатов: активные участники, кроме:
	// - старого ревьювера
	// - автора PR
	// - уже назначенных ревьюверов
	assignedMap := make(map[string]bool)
	for _, reviewerID := range pr.AssignedReviewers {
		assignedMap[reviewerID] = true
	}

	candidates := make([]entity.User, 0)
	for _, member := range teamMembers {
		if member.UserID != oldUserID &&
			member.UserID != pr.AuthorID &&
			!assignedMap[member.UserID] &&
			member.IsActive {
			candidates = append(candidates, member)
		}
	}

	// Проверяем наличие кандидатов
	if len(candidates) == 0 {
		s.log.Error("no candidates", zap.String("pr_id", pr.PullRequestID), zap.String("old_user_id", oldUserID))
		return nil, "", entity.ErrNoCandidate
	}

	// Выбираем случайного кандидата
	newReviewer := candidates[rand.Intn(len(candidates))]

	// Заменяем ревьювера
	if err := s.reviewerRepo.ReplaceReviewer(ctx, prID, oldUserID, newReviewer.UserID); err != nil {
		s.log.Error("replace reviewer", zap.Error(err))
		return nil, "", fmt.Errorf("replace reviewer: %w", err)
	}

	// Обновляем список ревьюверов в PR
	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldUserID {
			pr.AssignedReviewers[i] = newReviewer.UserID
			break
		}
	}
	if err := s.prRepo.Update(ctx, pr); err != nil {
		s.log.Error("update pr", zap.Error(err))
		return nil, "", fmt.Errorf("update pr: %w", err)
	}

	s.log.Info("reassigned reviewer", zap.String("pr_id", pr.PullRequestID), zap.String("old_user_id", oldUserID), zap.String("new_user_id", newReviewer.UserID))
	return pr, newReviewer.UserID, nil
}

// selectRandomReviewers выбирает случайных ревьюверов из списка кандидатов
func (s *PullRequestService) selectRandomReviewers(candidates []entity.User, maxCount int) []entity.User {
	if len(candidates) == 0 {
		s.log.Error("no candidates", zap.Int("max_count", maxCount))
		return []entity.User{}
	}

	count := min(maxCount, len(candidates))

	// Перемешиваем кандидатов
	shuffled := make([]entity.User, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	s.log.Info("selected reviewers", zap.Int("count", count))
	return shuffled[:count]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
