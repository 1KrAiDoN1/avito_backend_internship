package handler

import (
	"errors"
	"internship/internal/domain/entity"
	"internship/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService UserServiceInterface
	log         *zap.Logger
}

func NewUserHandler(userService UserServiceInterface, log *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// @Tags Users
// @Summary Установить флаг активности пользователя
func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req dto.SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			h.log.Error("user not found", zap.Error(err))
			respondError(c, http.StatusNotFound, entity.CodeNotFound, "user not found")
			return
		}
		h.log.Error("failed to update user", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to update user")
		return
	}

	h.log.Info("user updated", zap.Any("user", user))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// @Tags Users
// @Summary Получить PR'ы, где пользователь назначен ревьювером
func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		h.log.Error("user_id query parameter is required")
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "user_id query parameter is required")
		return
	}

	prs, err := h.userService.GetReviewPullRequests(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("failed to get review pull requests", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to get review pull requests")
		return
	}

	h.log.Info("review pull requests", zap.Any("prs", prs))
	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

// DeactivateTeamRequest представляет запрос на деактивацию команды
type DeactivateTeamRequest struct {
	TeamName string `json:"team_name" binding:"required"`
}

// @Tags Users
// @Summary Массово деактивировать участников команды
func (h *UserHandler) DeactivateTeam(c *gin.Context) {
	var req DeactivateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}

	affectedPRs, err := h.userService.DeactivateTeamMembers(c.Request.Context(), req.TeamName)
	if err != nil {
		h.log.Error("failed to deactivate team members", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to deactivate team members")
		return
	}

	h.log.Info("team members deactivated", zap.Any("affected_prs", affectedPRs))
	c.JSON(http.StatusOK, gin.H{
		"team_name":    req.TeamName,
		"affected_prs": affectedPRs,
		"message":      "Team members deactivated successfully",
	})
}
