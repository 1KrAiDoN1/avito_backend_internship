package handler

import (
	"errors"
	"fmt"
	"internship/domain/entity"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	teamService TeamServiceInterface
	log         *zap.Logger
}

func NewTeamHandler(teamService TeamServiceInterface, log *zap.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		log:         log,
	}
}

// @Tags Teams
// @Summary Создать команду с участниками (создаёт/обновляет пользователей)
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req entity.Team
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}

	if req.TeamName == "" {
		h.log.Error("team_name is required")
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "team_name is required")
		return
	}

	if len(req.Members) == 0 {
		h.log.Error("team must have at least one member")
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "team must have at least one member")
		return
	}
	exists, err := h.teamService.IsTeamExists(c.Request.Context(), req.TeamName)
	if err != nil {
		h.log.Error("failed to check team exists", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to check team exists")
		return
	}
	if exists {
		h.log.Error("team already exists", zap.String("team_name", req.TeamName))
		respondError(c, http.StatusBadRequest, entity.CodeTeamExists, fmt.Sprintf("team %s already exists", req.TeamName))
		return
	}

	createdTeam, err := h.teamService.CreateTeam(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("failed to create team", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to create team")
		return
	}
	h.log.Info("team created", zap.Any("team", createdTeam))
	c.JSON(http.StatusCreated, gin.H{"team": createdTeam})
}

// @Tags Teams
// @Summary Получить команду с участниками
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.log.Error("team_name query parameter is required")
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "team_name query parameter is required")
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if errors.Is(err, entity.ErrTeamNotFound) {
			h.log.Error("team not found")
			respondError(c, http.StatusNotFound, entity.CodeNotFound, "team not found")
			return
		}
		h.log.Error("failed to get team", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to get team")
		return
	}

	h.log.Info("team", zap.Any("team", team))
	c.JSON(http.StatusOK, gin.H{"team": team})
}
