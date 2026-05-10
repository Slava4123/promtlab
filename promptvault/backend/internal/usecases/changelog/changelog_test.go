// MN-1 — tests for changelog pure functions: computeHasUnread.
// Service.List/MarkSeen/HasUnread (с UserRepository mock) — отдельный PR.
package changelog

import (
	"testing"
	"time"
)

// computeHasUnread не требует репозитория — только хранит ссылку на embedded JSON.
// Создаём минимальный Service с заданным changelog для unit-теста.
func makeServiceWithDate(latestDate string) *Service {
	c := &Changelog{
		Entries: []Entry{
			{Date: latestDate, Title: "test"},
		},
	}
	return &Service{changelog: c}
}

func TestComputeHasUnread_NoLastSeen_True(t *testing.T) {
	svc := makeServiceWithDate("2026-05-01")
	if got := svc.computeHasUnread(nil); !got {
		t.Errorf("expected true when lastSeen=nil and entries exist, got false")
	}
}

func TestComputeHasUnread_LastSeenAfterLatest_False(t *testing.T) {
	svc := makeServiceWithDate("2026-05-01")
	lastSeen := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	if got := svc.computeHasUnread(&lastSeen); got {
		t.Errorf("expected false when lastSeen > latest, got true")
	}
}

func TestComputeHasUnread_LastSeenBeforeLatest_True(t *testing.T) {
	svc := makeServiceWithDate("2026-05-10")
	lastSeen := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	if got := svc.computeHasUnread(&lastSeen); !got {
		t.Errorf("expected true when lastSeen < latest, got false")
	}
}

func TestComputeHasUnread_EmptyChangelog_False(t *testing.T) {
	svc := &Service{changelog: &Changelog{Entries: nil}}
	lastSeen := time.Now()
	if got := svc.computeHasUnread(&lastSeen); got {
		t.Errorf("expected false for empty changelog, got true")
	}
}

func TestComputeHasUnread_InvalidDate_False(t *testing.T) {
	// При невалидной дате latestDate(), но непустой строке, мы доходим до
	// time.Parse → err → возвращаем false. nil lastSeen ловится раньше
	// (return true), поэтому передаём явный lastSeen != nil.
	svc := makeServiceWithDate("not-a-date")
	lastSeen := time.Now()
	if got := svc.computeHasUnread(&lastSeen); got {
		t.Errorf("expected false on unparseable date, got true (logged but no panic)")
	}
}

func TestLatestDate_Empty_ReturnsEmpty(t *testing.T) {
	svc := &Service{changelog: &Changelog{Entries: nil}}
	if got := svc.latestDate(); got != "" {
		t.Errorf("expected empty for no entries, got %q", got)
	}
}

func TestLatestDate_FirstEntry(t *testing.T) {
	svc := makeServiceWithDate("2026-01-15")
	if got := svc.latestDate(); got != "2026-01-15" {
		t.Errorf("expected 2026-01-15, got %q", got)
	}
}
