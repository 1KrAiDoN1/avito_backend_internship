package entity

import "errors"

// Доменные ошибки
var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
	ErrPRExists     = errors.New("pull request already exists")
	ErrPRNotFound   = errors.New("pull request not found")
	ErrPRMerged     = errors.New("cannot modify merged pull request")
	ErrNotAssigned  = errors.New("user is not assigned to this pull request")
	ErrNoCandidate  = errors.New("no active replacement candidate available")
	ErrInvalidInput = errors.New("invalid input data")
)

// ErrorCode представляет код ошибки API
type ErrorCode string

const (
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeNotFound    ErrorCode = "NOT_FOUND"
)

// APIError представляет структурированную ошибку API
type APIError struct {
	Code    ErrorCode //`json:"code"`
	Message string    //`json:"message"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error APIError //`json:"error"`
}
