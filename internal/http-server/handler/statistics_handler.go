package handler

import (
	"internship/internal/domain/entity"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatisticsHandler struct {
	statsService StatisticsServiceInterface
	log          *zap.Logger
}

func NewStatisticsHandler(statsService StatisticsServiceInterface, log *zap.Logger) *StatisticsHandler {
	return &StatisticsHandler{
		statsService: statsService,
		log:          log,
	}
}

// Получить статистику назначений и PR
func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
	stats, err := h.statsService.GetFullStats(c.Request.Context())
	if err != nil {
		h.log.Error("failed to get statistics", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to get statistics")
		return
	}

	h.log.Info("statistics", zap.Any("stats", stats))
	c.JSON(http.StatusOK, stats)
}
