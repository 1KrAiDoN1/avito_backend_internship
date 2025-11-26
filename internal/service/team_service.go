package service

import (
	"context"
	"fmt"
	"internship/internal/domain/entity"

	"go.uber.org/zap"
)

type TeamService struct {
	teamRepo TeamRepositoryInterface
	userRepo UserRepositoryInterface
	log      *zap.Logger
}

func NewTeamService(
	teamRepo TeamRepositoryInterface,
	userRepo UserRepositoryInterface,
	log *zap.Logger,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		log:      log,
	}
}

// CreateTeam создает команду и добавляет/обновляет участников
func (s *TeamService) CreateTeam(ctx context.Context, team *entity.Team) (*entity.Team, error) {

	if err := s.teamRepo.Create(ctx, team); err != nil {
		s.log.Error("create team", zap.Error(err))
		return nil, fmt.Errorf("create team: %w", err)
	}
	users := make([]*entity.User, 0, len(team.Members))
	for _, member := range team.Members {
		users = append(users, &entity.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		})
	}
	if err := s.userRepo.BatchCreateOrUpdate(ctx, users); err != nil {
		s.log.Error("batch create or update users", zap.Error(err))
		return nil, fmt.Errorf("batch create or update users: %w", err)
	}

	return team, nil
}

// IsTeamExists проверяет существование команды
func (s *TeamService) IsTeamExists(ctx context.Context, teamName string) (bool, error) {
	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		s.log.Error("check team exists", zap.Error(err))
		return false, fmt.Errorf("check team exists: %w", err)
	}

	return exists, nil
}

// GetTeam получает команду с участниками
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		s.log.Error("get team", zap.Error(err))
		return nil, fmt.Errorf("get team: %w", err)
	}

	users, err := s.userRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		s.log.Error("get team members", zap.Error(err))
		return nil, fmt.Errorf("get team members: %w", err)
	}

	team.Members = make([]entity.TeamMember, 0, len(users))
	for _, user := range users {
		team.Members = append(team.Members, entity.TeamMember{
			UserID:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		})
	}

	s.log.Info("team", zap.Any("team", team))
	return team, nil
}
