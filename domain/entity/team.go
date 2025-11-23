package entity

// Team представляет команду разработчиков
type Team struct {
	TeamName string       `json:"team_name" db:"team_name"`
	Members  []TeamMember `json:"members" db:"-"`
}

// TeamMember представляет участника команды в составе команды
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
