package analytics

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/pkg/safeloop"
)

// insightsParallelism — кол-во concurrent ComputeInsights worker'ов в loop'е.
// MN-46: раньше per-user serial — на 50 Max-юзерах × ~2с/юзер = 100с.
// 4 worker'а → ~25с на ту же выборку, при этом PG не задыхается (на VPS 4GB
// pool=15 connections, оставляем запас на active webhook'и и API).
const insightsParallelism = 4

// InsightsComputeLoop — ежесуточный пересчёт Smart Insights для платных юзеров.
// Идёт по списку активных Pro/Max-подписчиков (users.plan_id IN (pro,
// pro_yearly, max, max_yearly), status='active') и вызывает
// Service.ComputeInsights для каждого (personal scope + по каждой команде,
// которой он владеет). Per-plan dispatch (Pro → 2 teaser типа, Max → все 7)
// делается через GetByID + insightsForPlan на каждой итерации.
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

// compute — итерация по платным юзерам через UserRepository.ListPaidUsers.
// Для каждого: 1) personal scope; 2) для каждой команды, где он owner — team scope.
// Если запрос ListPaidUsers фейлится — единственный slog.Error, loop продолжит завтра.
//
// Для MVP: один batch без пагинации. При росте платных юзеров (>1000) —
// добавить пагинацию и распределение по окну, чтобы не долбить БД одновременно.
func (l *InsightsComputeLoop) compute() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	metrics.InsightsLoopRuns.Inc()

	ids, err := l.users.ListPaidUsers(ctx)
	if err != nil {
		slog.Error("analytics.insights_loop.list_failed", "error", err)
		metrics.InsightsRefresh.WithLabelValues("error").Inc()
		return
	}
	// MN-46: errgroup с лимитом insightsParallelism — параллельно считаем
	// несколько платных юзеров. Atomic-счётчики защищают агрегаты от data race.
	var (
		okCount   atomic.Int64
		failCount atomic.Int64
		teamOk    atomic.Int64
		teamFail  atomic.Int64
		skipCount atomic.Int64 // race: юзер изменил plan между snapshot и compute
		mu        sync.Mutex   // защита для slog.Warn вывода
	)
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(insightsParallelism)
	for _, uid := range ids {
		uid := uid
		g.Go(func() error {
			// Task 7 (Pricing Iteration v3): per-plan dispatch. Делаем extra
			// DB-roundtrip GetByID(uid) чтобы прочитать актуальный PlanID
			// и передать корректный allowed-набор в ComputeInsights.
			//
			// Цена roundtrip'а: ListPaidUsers ограничивает выборку платными
			// юзерами (single SELECT by PK, кеш-friendly), parallelism=4
			// амортизирует latency.
			//
			// Race window: между ListPaidUsers (snapshot) и GetByID (current)
			// юзер мог downgrade'нуть на Free / переключиться на yearly.
			// insightsForPlan вернёт nil для Free → allowed пуст → skip без
			// compute. Для других paid plan'ов отработает корректно.
			user, gerr := l.users.GetByID(gctx, uid)
			if gerr != nil {
				if errors.Is(gerr, repo.ErrNotFound) {
					// Юзер удалён между snapshot и compute — skip тихо.
					skipCount.Add(1)
					return nil
				}
				mu.Lock()
				slog.WarnContext(gctx, "analytics.insights_loop.get_user_failed",
					"user_id", uid, "error", gerr)
				mu.Unlock()
				failCount.Add(1)
				metrics.InsightsRefresh.WithLabelValues("error").Inc()
				return nil
			}
			allowed := l.svc.insightsForPlan(user.PlanID)
			if len(allowed) == 0 {
				// Race: юзер на Free/unknown plan (downgrade между ListPaidUsers
				// и GetByID), либо Pro при выключенном PRO_INSIGHTS_TEASER_ENABLED.
				// Skip без compute — ComputeInsights бы no-op'нул, но мы избегаем
				// лишних SQL-запросов в svc.
				skipCount.Add(1)
				return nil
			}

			// 1. Personal scope.
			if cerr := l.svc.ComputeInsights(gctx, uid, nil, allowed); cerr != nil {
				failCount.Add(1)
				metrics.InsightsRefresh.WithLabelValues("error").Inc()
			} else {
				okCount.Add(1)
				metrics.InsightsRefresh.WithLabelValues("success").Inc()
			}

			// 2. Team scope — только команды, где юзер owner.
			// Используем тот же allowed-набор (per-юзера, не per-team):
			// owner получает на team-scope то же, что и на personal.
			teams, terr := l.teams.ListOwnedTeams(gctx, uid)
			if terr != nil {
				mu.Lock()
				slog.WarnContext(gctx, "analytics.insights_loop.list_owned_teams_failed",
					"user_id", uid, "error", terr)
				mu.Unlock()
				metrics.InsightsTeamRun.WithLabelValues("error").Inc()
				return nil
			}
			for _, team := range teams {
				tid := team.ID
				if cerr := l.svc.ComputeInsights(gctx, uid, &tid, allowed); cerr != nil {
					mu.Lock()
					slog.WarnContext(gctx, "analytics.insights_loop.team_compute_failed",
						"user_id", uid, "team_id", tid, "error", cerr)
					mu.Unlock()
					metrics.InsightsTeamRun.WithLabelValues("error").Inc()
					teamFail.Add(1)
				} else {
					metrics.InsightsTeamRun.WithLabelValues("success").Inc()
					teamOk.Add(1)
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		// errgroup отлавливает первую ошибку из горутин (panic/ctx cancel/
		// pool exhaustion). Раньше эту ошибку молча проглатывали через
		// `_ = g.Wait()`, и operators читали "ok=0 failed=0 total=N" как
		// "нечего обрабатывать" вместо "всё упало".
		slog.Error("analytics.insights_loop.run.wait_failed", "err", err)
	}
	slog.Info("analytics.insights_loop.run",
		"ok", okCount.Load(), "failed", failCount.Load(), "total", len(ids),
		"skipped", skipCount.Load(),
		"team_ok", teamOk.Load(), "team_failed", teamFail.Load())
}
