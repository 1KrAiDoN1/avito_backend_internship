package handler

import "go.uber.org/zap"

type Handlers struct {
	TeamHandler        *TeamHandler
	UserHandler        *UserHandler
	PullRequestHandler *PullRequestHandler
	StatisticsHandler  *StatisticsHandler
}

func NewHandlers(teamService TeamServiceInterface, userService UserServiceInterface, pullRequestService PullRequestServiceInterface, statisticsService StatisticsServiceInterface, log *zap.Logger) *Handlers {
	return &Handlers{
		TeamHandler:        NewTeamHandler(teamService, log),
		UserHandler:        NewUserHandler(userService, log),
		PullRequestHandler: NewPullRequestHandler(pullRequestService, log),
		StatisticsHandler:  NewStatisticsHandler(statisticsService, log),
	}
}
