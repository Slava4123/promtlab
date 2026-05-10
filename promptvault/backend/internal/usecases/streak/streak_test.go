// MN-1 — pure-function test for streak.todayInTimezone.
// Service.RecordActivity / GetStreak требуют StreakRepository mock — отдельный PR.
package streak

import (
	"strings"
	"testing"
	"time"
)

func TestTodayInTimezone_Empty_UsesUTC(t *testing.T) {
	got := todayInTimezone("")
	want := time.Now().UTC().Format("2006-01-02")
	if got != want {
		t.Errorf("todayInTimezone('') = %q, want %q (UTC today)", got, want)
	}
}

func TestTodayInTimezone_ValidTZ(t *testing.T) {
	got := todayInTimezone("Europe/Moscow")
	if !validDate(got) {
		t.Errorf("todayInTimezone('Europe/Moscow') = %q, expected YYYY-MM-DD", got)
	}
	loc, _ := time.LoadLocation("Europe/Moscow")
	want := time.Now().In(loc).Format("2006-01-02")
	if got != want {
		t.Errorf("todayInTimezone('Europe/Moscow') = %q, want %q", got, want)
	}
}

func TestTodayInTimezone_InvalidTZ_FallbackUTC(t *testing.T) {
	got := todayInTimezone("Bogus/Timezone")
	want := time.Now().UTC().Format("2006-01-02")
	if got != want {
		t.Errorf("todayInTimezone('Bogus/Timezone') = %q, want %q (fallback UTC)", got, want)
	}
}

func validDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	if !strings.Contains(s, "-") {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
