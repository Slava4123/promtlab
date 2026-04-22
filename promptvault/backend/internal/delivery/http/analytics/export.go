package analytics

import (
	"fmt"
	"log/slog"
	"mime"
	"net/http"

	"github.com/xuri/excelize/v2"

	repo "promptvault/internal/interface/repository"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// writeUsageCSV — CSV-export в один sheet (date,uses). Зеркало старого
// поведения, чтобы не ломать пользователей которые уже скриптами парсят
// формат "date,uses" без header-строк.
func writeUsageCSV(w http.ResponseWriter, baseName string, points []repo.UsagePoint) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition",
		mime.FormatMediaType("attachment", map[string]string{"filename": baseName + ".csv"}))
	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprintln(w, "date,uses"); err != nil {
		return
	}
	for _, p := range points {
		if _, err := fmt.Fprintf(w, "%s,%d\n", p.Day.Format("2006-01-02"), p.Count); err != nil {
			return
		}
	}
}

// writePersonalXLSX — 4 sheet'а: Usage, Top Prompts, Top Shared, By Model.
// Возвращает файл как application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.
func writePersonalXLSX(w http.ResponseWriter, baseName string, d *analyticsuc.PersonalDashboard) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	writeUsageSheet(f, "Использование", d.UsagePerDay)
	writeTopPromptsSheet(f, "Топ промптов", d.TopPrompts, "Использований")
	writeTopPromptsSheet(f, "Топ просмотров ссылок", d.TopShared, "Просмотров")
	writeModelSheet(f, "По моделям", d.UsageByModel)

	// Excel создаёт дефолтный "Sheet1" — убираем после добавления наших.
	_ = f.DeleteSheet("Sheet1")

	finalizeXLSX(w, baseName, f)
}

// writeTeamXLSX — 4 sheet'а: Usage, Top Prompts, Contributors, By Model.
func writeTeamXLSX(w http.ResponseWriter, baseName string, d *analyticsuc.TeamDashboard) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	writeUsageSheet(f, "Использование", d.UsagePerDay)
	writeTopPromptsSheet(f, "Топ промптов", d.TopPrompts, "Использований")
	writeContributorsSheet(f, "Вклад участников", d.Contributors)
	writeModelSheet(f, "По моделям", d.UsageByModel)

	_ = f.DeleteSheet("Sheet1")

	finalizeXLSX(w, baseName, f)
}

func finalizeXLSX(w http.ResponseWriter, baseName string, f *excelize.File) {
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition",
		mime.FormatMediaType("attachment", map[string]string{"filename": baseName + ".xlsx"}))
	w.WriteHeader(http.StatusOK)
	if err := f.Write(w); err != nil {
		slog.Warn("analytics.export.xlsx.write_failed", "err", err)
	}
}

func writeUsageSheet(f *excelize.File, name string, points []repo.UsagePoint) {
	if _, err := f.NewSheet(name); err != nil {
		slog.Warn("xlsx.new_sheet", "err", err, "sheet", name)
		return
	}
	_ = f.SetCellValue(name, "A1", "Дата")
	_ = f.SetCellValue(name, "B1", "Использований")
	for i, p := range points {
		row := i + 2
		_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), p.Day.Format("2006-01-02"))
		_ = f.SetCellValue(name, fmt.Sprintf("B%d", row), p.Count)
	}
}

func writeTopPromptsSheet(f *excelize.File, name string, rows []repo.PromptUsageRow, metricLabel string) {
	if _, err := f.NewSheet(name); err != nil {
		slog.Warn("xlsx.new_sheet", "err", err, "sheet", name)
		return
	}
	_ = f.SetCellValue(name, "A1", "ID промпта")
	_ = f.SetCellValue(name, "B1", "Название")
	_ = f.SetCellValue(name, "C1", metricLabel)
	for i, r := range rows {
		row := i + 2
		_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), r.PromptID)
		_ = f.SetCellValue(name, fmt.Sprintf("B%d", row), r.Title)
		_ = f.SetCellValue(name, fmt.Sprintf("C%d", row), r.Uses)
	}
}

func writeContributorsSheet(f *excelize.File, name string, rows []repo.ContributorRow) {
	if _, err := f.NewSheet(name); err != nil {
		slog.Warn("xlsx.new_sheet", "err", err, "sheet", name)
		return
	}
	_ = f.SetCellValue(name, "A1", "Email")
	_ = f.SetCellValue(name, "B1", "Имя")
	_ = f.SetCellValue(name, "C1", "Создано")
	_ = f.SetCellValue(name, "D1", "Обновлений")
	_ = f.SetCellValue(name, "E1", "Использований")
	for i, r := range rows {
		row := i + 2
		_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), r.Email)
		_ = f.SetCellValue(name, fmt.Sprintf("B%d", row), r.Name)
		_ = f.SetCellValue(name, fmt.Sprintf("C%d", row), r.PromptsCreated)
		_ = f.SetCellValue(name, fmt.Sprintf("D%d", row), r.PromptsEdited)
		_ = f.SetCellValue(name, fmt.Sprintf("E%d", row), r.Uses)
	}
}

func writeModelSheet(f *excelize.File, name string, rows []repo.ModelUsageRow) {
	if _, err := f.NewSheet(name); err != nil {
		slog.Warn("xlsx.new_sheet", "err", err, "sheet", name)
		return
	}
	_ = f.SetCellValue(name, "A1", "Модель")
	_ = f.SetCellValue(name, "B1", "Использований")
	for i, r := range rows {
		row := i + 2
		label := r.Model
		if label == "" {
			label = "Без модели"
		}
		_ = f.SetCellValue(name, fmt.Sprintf("A%d", row), label)
		_ = f.SetCellValue(name, fmt.Sprintf("B%d", row), r.Uses)
	}
}
