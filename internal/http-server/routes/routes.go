package routes

import (
	"internship/internal/http-server/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(
	router *gin.RouterGroup,
	logger *zap.Logger,
	handlers *handler.Handlers,
) {

	team := router.Group("/team")
	{
		team.GET("/get", handlers.TeamHandler.GetTeam)
		team.POST("/add", handlers.TeamHandler.CreateTeam)
	}

	users := router.Group("/users")
	{
		users.POST("/setIsActive", handlers.UserHandler.SetIsActive)
		users.GET("/getReview", handlers.UserHandler.GetReview)
		users.POST("/deactivateTeam", handlers.UserHandler.DeactivateTeam)
	}

	pullRequests := router.Group("/pullRequests")
	{
		pullRequests.POST("/create", handlers.PullRequestHandler.CreatePullRequest)
		pullRequests.POST("/merge", handlers.PullRequestHandler.MergePullRequest)
		pullRequests.POST("/reassign", handlers.PullRequestHandler.ReassignReviewer)
	}

	statistics := router.Group("/statistics")
	{
		statistics.GET("", handlers.StatisticsHandler.GetStatistics)
	}

}
