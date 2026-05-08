package analytics

import (
	"context"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/pkg/safeloop"
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

	safeloop.RunWithRecover("insights_compute", l.compute)
	for {
		select {
		case <-ticker.C:
			safeloop.RunWithRecover("insights_compute", l.compute)
		case <-l.stopCh:
			return
		}
	}
}

// compute — итерация по Max-юзерам через UserRepository.ListMaxUsers.
// Для каждого: 1) personal scope; 2) для каждой команды, где он owner — team scope.
// Если запрос ListMaxUsers фейлится — единственный slog.Error, loop продолжит завтра.
//
// Для MVP: один batch без пагинации. При росте Max-юзеров (>1000) — добавить
// пагинацию и распределение по окну, чтобы не долбить БД одновременно.
func (l *InsightsComputeLoop) compute() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	metrics.InsightsLoopRuns.Inc()

	ids, err := l.users.ListMaxUsers(ctx)
	if err != nil {
		slog.Error("analytics.insights_loop.list_failed", "error", err)
		metrics.InsightsRefresh.WithLabelValues("error").Inc()
		return
	}
	var okCount, failCount, teamOk, teamFail int
	for _, uid := range ids {
		// 1. Personal scope.
		if cerr := l.svc.ComputeInsights(ctx, uid, nil); cerr != nil {
			failCount++
			metrics.InsightsRefresh.WithLabelValues("error").Inc()
		} else {
			okCount++
			metrics.InsightsRefresh.WithLabelValues("success").Inc()
		}

		// 2. Team scope — только команды, где юзер owner.
		teams, terr := l.teams.ListOwnedTeams(ctx, uid)
		if terr != nil {
			slog.WarnContext(ctx, "analytics.insights_loop.list_owned_teams_failed",
				"user_id", uid, "error", terr)
			metrics.InsightsTeamRun.WithLabelValues("error").Inc()
			continue
		}
		for _, team := range teams {
			tid := team.ID
			if cerr := l.svc.ComputeInsights(ctx, uid, &tid); cerr != nil {
				slog.WarnContext(ctx, "analytics.insights_loop.team_compute_failed",
					"user_id", uid, "team_id", tid, "error", cerr)
				metrics.InsightsTeamRun.WithLabelValues("error").Inc()
				teamFail++
			} else {
				metrics.InsightsTeamRun.WithLabelValues("success").Inc()
				teamOk++
			}
		}
	}
	slog.Info("analytics.insights_loop.run",
		"ok", okCount, "failed", failCount, "total", len(ids),
		"team_ok", teamOk, "team_failed", teamFail)
}
