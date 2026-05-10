package models

import "time"

type UserStreak struct {
	UserID        uint      `gorm:"primaryKey" json:"user_id"`
	CurrentStreak int       `gorm:"not null;default:0" json:"current_streak"`
	LongestStreak int       `gorm:"not null;default:0" json:"longest_streak"`
	// MJ-29: было `string \`type:date\``. Сравнение `streak.LastActiveDate == today.Format(...)`
	// хрупко зависело от драйвера (lib/pq возвращал YYYY-MM-DD, pgx — с timezone offset).
	// Теперь — типизированный date через *time.Time (как QuotaWarningSentOn в users).
	LastActiveDate *time.Time `gorm:"type:date;not null" json:"-"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// ReminderSentOn — M-16: дата, когда в последний раз отправляли "не сломай серию"
	// напоминание. Защита от дубликата при повторном тике loop в тот же день.
	ReminderSentOn *time.Time `gorm:"column:reminder_sent_on;type:date" json:"-"`
}

// LastActiveDateString возвращает дату как YYYY-MM-DD для JSON API совместимости.
// JSON-tag `last_active_date` сериализуется через MarshalJSON ниже.
func (u UserStreak) LastActiveDateString() string {
	if u.LastActiveDate == nil {
		return ""
	}
	return u.LastActiveDate.Format("2006-01-02")
}

func (UserStreak) TableName() string {
	return "user_streaks"
}
