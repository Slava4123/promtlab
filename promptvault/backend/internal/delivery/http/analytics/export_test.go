package analytics

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"

	repo "promptvault/internal/interface/repository"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// Регрессия на B.6 (xlsx multi-sheet export): проверяем что файл
// генерируется без паники, содержит 4 sheet'а и заполнены данные.

func TestWritePersonalXLSX_ContainsFourSheets(t *testing.T) {
	rec := httptest.NewRecorder()
	day := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	d := &analyticsuc.PersonalDashboard{
		Range:       analyticsuc.Range7d,
		UsagePerDay: []repo.UsagePoint{{Day: day, Count: 42}},
		TopPrompts:  []repo.PromptUsageRow{{PromptID: 1, Title: "My Prompt", Uses: 99}},
		TopShared:   []repo.PromptUsageRow{{PromptID: 2, Title: "Shared", Uses: 17}},
		UsageByModel: []repo.ModelUsageRow{
			{Model: "claude-sonnet-4", Uses: 30},
			{Model: "", Uses: 12},
		},
	}

	writePersonalXLSX(rec, "analytics-personal-7d", d)

	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), ".xlsx")

	// Открываем ответ как xlsx, проверяем sheet'ы и content.
	f, err := excelize.OpenReader(bytes.NewReader(rec.Body.Bytes()))
	if !assert.NoError(t, err, "response must be valid xlsx") {
		return
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	assert.Contains(t, sheets, "Использование")
	assert.Contains(t, sheets, "Топ промптов")
	assert.Contains(t, sheets, "Топ просмотров ссылок")
	assert.Contains(t, sheets, "По моделям")
	assert.NotContains(t, sheets, "Sheet1", "дефолтный Sheet1 должен быть удалён")

	// Headers первого sheet.
	h1, _ := f.GetCellValue("Использование", "A1")
	h2, _ := f.GetCellValue("Использование", "B1")
	assert.Equal(t, "Дата", h1)
	assert.Equal(t, "Использований", h2)

	// Первая строка данных — дата + count=42.
	v, _ := f.GetCellValue("Использование", "A2")
	assert.Equal(t, "2026-04-20", v)
	count, _ := f.GetCellValue("Использование", "B2")
	assert.Equal(t, "42", count)

	// В "По моделям" — пустая модель должна быть заменена на «Без модели».
	modelCell, _ := f.GetCellValue("По моделям", "A3")
	assert.Equal(t, "Без модели", modelCell)
}

func TestWriteTeamXLSX_HasContributorsSheet(t *testing.T) {
	rec := httptest.NewRecorder()
	d := &analyticsuc.TeamDashboard{
		Range: analyticsuc.Range30d,
		Contributors: []repo.ContributorRow{
			{UserID: 1, Email: "a@example.com", Name: "Alice", PromptsCreated: 3, PromptsEdited: 1, Uses: 20},
			{UserID: 2, Email: "b@example.com", Name: "Bob", PromptsCreated: 0, PromptsEdited: 2, Uses: 5},
		},
	}

	writeTeamXLSX(rec, "analytics-team-5-30d", d)

	f, err := excelize.OpenReader(bytes.NewReader(rec.Body.Bytes()))
	if !assert.NoError(t, err) {
		return
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	assert.Contains(t, sheets, "Вклад участников")

	// Первая строка контрибьютора.
	email, _ := f.GetCellValue("Вклад участников", "A2")
	name, _ := f.GetCellValue("Вклад участников", "B2")
	uses, _ := f.GetCellValue("Вклад участников", "E2")
	assert.Equal(t, "a@example.com", email)
	assert.Equal(t, "Alice", name)
	assert.Equal(t, "20", uses)
}

func TestWriteUsageCSV_OnlyDateUsesFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	day := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	points := []repo.UsagePoint{
		{Day: day, Count: 7},
		{Day: day.AddDate(0, 0, 1), Count: 3},
	}

	writeUsageCSV(rec, "test", points)

	body := rec.Body.String()
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, body, "date,uses")
	assert.Contains(t, body, "2026-04-20,7")
	assert.Contains(t, body, "2026-04-21,3")
}
