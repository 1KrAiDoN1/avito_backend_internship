package handler

import (
	"internship/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

// respondError отправляет стандартизированный ответ с ошибкой
func respondError(c *gin.Context, statusCode int, code entity.ErrorCode, message string) {
	c.JSON(statusCode, entity.ErrorResponse{
		Error: entity.APIError{
			Code:    code,
			Message: message,
		},
	})
}
