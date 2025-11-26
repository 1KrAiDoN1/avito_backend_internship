package handler

import (
	"errors"
	"internship/internal/domain/entity"
	"internship/internal/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PullRequestHandler struct {
	prService PullRequestServiceInterface
	log       *zap.Logger
}

func NewPullRequestHandler(prService PullRequestServiceInterface, log *zap.Logger) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
		log:       log,
	}
}

// @Tags PullRequests
// @SummaryСоздать PR и автоматически назначить до 2 ревьюверов из команды автора
func (h *PullRequestHandler) CreatePullRequest(c *gin.Context) {
	var req dto.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}
	pr := &entity.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}
	createdPR, err := h.prService.CreatePullRequest(c.Request.Context(), pr)
	if err != nil {
		if errors.Is(err, entity.ErrPRExists) {
			h.log.Error("PR id already exists", zap.Error(err))
			respondError(c, http.StatusConflict, entity.CodePRExists, "PR id already exists")
			return
		}
		if errors.Is(err, entity.ErrUserNotFound) {
			h.log.Error("author not found", zap.Error(err))
			respondError(c, http.StatusNotFound, entity.CodeNotFound, "author not found")
			return
		}
		h.log.Error("failed to create pull request", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to create pull request")
		return
	}

	h.log.Info("pull request created", zap.Any("pr", createdPR))
	c.JSON(http.StatusCreated, gin.H{"pr": createdPR})
}

// @Tags PullRequests
// @Summary Пометить PR как MERGED (идемпотентная операция)
func (h *PullRequestHandler) MergePullRequest(c *gin.Context) {
	var req dto.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}

	pr, err := h.prService.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, entity.ErrPRNotFound) {
			h.log.Error("pull request not found", zap.Error(err))
			respondError(c, http.StatusNotFound, entity.CodeNotFound, "pull request not found")
			return
		}
		h.log.Error("failed to merge pull request", zap.Error(err))
		respondError(c, http.StatusInternalServerError, entity.CodeNotFound, "failed to merge pull request")
		return
	}

	h.log.Info("pull request merged", zap.Any("pr", pr))
	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

// @Tags PullRequests
// @Summary Переназначить конкретного ревьювера на другого из его команды
func (h *PullRequestHandler) ReassignReviewer(c *gin.Context) {
	var req dto.ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, entity.CodeNotFound, "invalid request body")
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(
		c.Request.Context(),
		req.PullRequestID,
		req.OldUserID,
	)

	if err != nil {
		h.log.Error("reassign reviewer", zap.Error(err))
		statusCode := http.StatusInternalServerError
		code := entity.CodeNotFound

		switch {
		case errors.Is(err, entity.ErrPRNotFound), errors.Is(err, entity.ErrUserNotFound):
			statusCode = http.StatusNotFound
			code = entity.CodeNotFound
		case errors.Is(err, entity.ErrPRMerged):
			statusCode = http.StatusConflict
			code = entity.CodePRMerged
		case errors.Is(err, entity.ErrNotAssigned):
			statusCode = http.StatusConflict
			code = entity.CodeNotAssigned
		case errors.Is(err, entity.ErrNoCandidate):
			statusCode = http.StatusConflict
			code = entity.CodeNoCandidate
		}

		h.log.Error("reassign reviewer", zap.Error(err))
		respondError(c, statusCode, code, err.Error())
		return
	}

	h.log.Info("reassign reviewer", zap.Any("pr", pr), zap.String("new_reviewer_id", newReviewerID))
	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
