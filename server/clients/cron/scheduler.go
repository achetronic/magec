package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

// Scheduler runs cron clients on their configured schedules.
// It polls the store periodically for client changes rather than
// depending on a cron library, keeping dependencies minimal.
type Scheduler struct {
	executor *clients.Executor
	store    *store.Store
	logger   *slog.Logger

	mu      sync.Mutex
	cancel  context.CancelFunc
	entries map[string]*cronEntry
}

type cronEntry struct {
	client   store.ClientDefinition
	schedule *Schedule
	next     time.Time
}

// NewScheduler creates a cron scheduler that checks clients every 30 seconds.
func NewScheduler(executor *clients.Executor, s *store.Store, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		executor: executor,
		store:    s,
		logger:   logger,
		entries:  make(map[string]*cronEntry),
	}
}

// Start begins the scheduler loop. It reloads cron clients from the store
// whenever they change and fires matching entries every 30 seconds.
func (s *Scheduler) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	s.reload()

	changeCh := s.store.OnChange()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-changeCh:
			s.reload()
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Scheduler) reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	allClients := s.store.ListClients()
	newEntries := make(map[string]*cronEntry, len(allClients))

	for _, cl := range allClients {
		if cl.Type != "cron" || !cl.Enabled || cl.Config.Cron == nil {
			continue
		}

		sched, err := Parse(cl.Config.Cron.Schedule)
		if err != nil {
			s.logger.Warn("Invalid cron schedule, skipping", "client", cl.Name, "schedule", cl.Config.Cron.Schedule, "error", err)
			continue
		}

		if existing, ok := s.entries[cl.ID]; ok {
			existing.client = cl
			existing.schedule = sched
			newEntries[cl.ID] = existing
		} else {
			newEntries[cl.ID] = &cronEntry{
				client:   cl,
				schedule: sched,
				next:     sched.Next(time.Now()),
			}
		}
	}

	s.entries = newEntries
	s.logger.Debug("Scheduler reloaded", "cronClients", len(newEntries))
}

func (s *Scheduler) tick(ctx context.Context) {
	s.mu.Lock()
	now := time.Now()
	var toRun []cronEntry
	for _, entry := range s.entries {
		if !now.Before(entry.next) {
			toRun = append(toRun, *entry)
			entry.next = entry.schedule.Next(now)
		}
	}
	s.mu.Unlock()

	for _, entry := range toRun {
		go func(cl store.ClientDefinition) {
			s.logger.Info("Cron client firing", "client", cl.Name, "id", cl.ID)
			result, err := s.executor.RunClient(ctx, cl, "")
			if err != nil {
				s.logger.Error("Cron client failed", "client", cl.Name, "error", err)
				return
			}
			s.logger.Info("Cron client completed", "client", cl.Name, "responseLen", len(result))
		}(entry.client)
	}
}
