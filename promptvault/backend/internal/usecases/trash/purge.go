package trash

import (
	"context"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

type PurgeLoop struct {
	repo      repo.TrashRepository
	interval  time.Duration
	retention int
	stopCh    chan struct{}
}

func NewPurgeLoop(r repo.TrashRepository, interval time.Duration, retentionDays int) *PurgeLoop {
	return &PurgeLoop{
		repo:      r,
		interval:  interval,
		retention: retentionDays,
		stopCh:    make(chan struct{}),
	}
}

func (p *PurgeLoop) Start() {
	go p.run()
}

func (p *PurgeLoop) Stop() {
	close(p.stopCh)
}

func (p *PurgeLoop) run() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Первый запуск сразу при старте
	p.purge()

	for {
		select {
		case <-ticker.C:
			p.purge()
		case <-p.stopCh:
			return
		}
	}
}

func (p *PurgeLoop) purge() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deleted, err := p.repo.PurgeExpired(ctx, p.retention)
	if err != nil {
		slog.Error("trash.purge.failed", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("trash.purge", "deleted", deleted, "retention_days", p.retention)
	}
}
