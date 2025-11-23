package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type StatisticsService struct {
	statsRepo StatisticsRepositoryInterface
	log       *zap.Logger
}

// NewStatisticsService создает новый сервис статистики
func NewStatisticsService(statsRepo StatisticsRepositoryInterface, log *zap.Logger) *StatisticsService {
	return &StatisticsService{
		statsRepo: statsRepo,
		log:       log,
	}
}

// GetAssignmentStats возвращает статистику назначений по пользователям
func (s *StatisticsService) GetAssignmentStats(ctx context.Context) (map[string]int, error) {
	stats, err := s.statsRepo.GetAssignmentStats(ctx)
	if err != nil {
		s.log.Error("get assignment stats", zap.Error(err))
		return nil, fmt.Errorf("get assignment stats: %w", err)
	}
	return stats, nil
}

// GetPRStats возвращает статистику по PR
func (s *StatisticsService) GetPRStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.statsRepo.GetPRStats(ctx)
	if err != nil {
		s.log.Error("get pr stats", zap.Error(err))
		return nil, fmt.Errorf("get pr stats: %w", err)
	}
	return stats, nil
}

// GetFullStats возвращает полную статистику
func (s *StatisticsService) GetFullStats(ctx context.Context) (map[string]interface{}, error) {
	assignmentStats, err := s.GetAssignmentStats(ctx)
	if err != nil {
		s.log.Error("get assignment stats", zap.Error(err))
		return nil, fmt.Errorf("get assignment stats: %w", err)
	}

	prStats, err := s.GetPRStats(ctx)
	if err != nil {
		s.log.Error("get pr stats", zap.Error(err))
		return nil, fmt.Errorf("get pr stats: %w", err)
	}

	fullStats := map[string]interface{}{
		"assignments_by_user": assignmentStats,
		"pull_requests":       prStats,
	}

	s.log.Info("full stats", zap.Any("full_stats", fullStats))
	return fullStats, nil
}
