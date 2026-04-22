package analytics

import (
	"context"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

// InsightsComputeLoop — ежесуточный пересчёт Smart Insights для Max-юзеров.
// Идёт по списку активных Max-подписчиков (users.plan_id LIKE 'max%') и
// вызывает Service.ComputeInsights для каждого (personal scope + по
// каждой команде, которой он владеет).
//
// Для MVP: итерация в один batch. При росте users'ов — пагинация и
// распределение по окну (чтобы не долбить БД в одну секунду).
type InsightsComputeLoop struct {
	svc      *Service
	users    repo.UserRepository
	teams    repo.TeamRepository
	interval time.Duration
	stopCh   chan struct{}
}

func NewInsightsComputeLoop(svc *Service, users repo.UserRepository, teams repo.TeamRepository, interval time.Duration) *InsightsComputeLoop {
	return &InsightsComputeLoop{
		svc:      svc,
		users:    users,
		teams:    teams,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (l *InsightsComputeLoop) Start() { go l.run() }
func (l *InsightsComputeLoop) Stop()  { close(l.stopCh) }

func (l *InsightsComputeLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	l.compute()
	for {
		select {
		case <-ticker.C:
			l.compute()
		case <-l.stopCh:
			return
		}
	}
}

// compute — итерация по Max-юзерам через UserRepository.ListMaxUsers.
// Если запрос фейлится, единственный slog.Error — loop продолжит завтра.
func (l *InsightsComputeLoop) compute() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ids, err := l.users.ListMaxUsers(ctx)
	if err != nil {
		slog.Error("analytics.insights_loop.list_failed", "error", err)
		return
	}
	var okCount, failCount int
	for _, uid := range ids {
		if cerr := l.svc.ComputeInsights(ctx, uid, nil); cerr != nil {
			failCount++
			continue
		}
		okCount++
		// Персональный scope посчитан. Для команд владельца — отдельный проход
		// потребует TeamRepository.ListByOwnerID, который пока отсутствует;
		// TODO: расширить, когда Max-юзер с командой реально запросит.
	}
	slog.Info("analytics.insights_loop.run", "ok", okCount, "failed", failCount, "total", len(ids))
}
