package streak

type StreakOutput struct {
	CurrentStreak  int    `json:"current_streak"`
	LongestStreak  int    `json:"longest_streak"`
	LastActiveDate string `json:"last_active_date"`
	ActiveToday    bool   `json:"active_today"`
}
