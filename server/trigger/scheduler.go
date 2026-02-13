package trigger

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/achetronic/magec/server/store"
)

// Scheduler runs cron triggers on their configured schedules.
// It polls the store periodically for trigger changes rather than
// depending on a cron library, keeping dependencies minimal.
type Scheduler struct {
	executor *Executor
	store    *store.Store
	logger   *slog.Logger

	mu      sync.Mutex
	cancel  context.CancelFunc
	entries map[string]*cronEntry
}

type cronEntry struct {
	trigger  store.Trigger
	schedule *cronSchedule
	next     time.Time
}

// NewScheduler creates a cron scheduler that checks triggers every minute.
func NewScheduler(executor *Executor, s *store.Store, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		executor: executor,
		store:    s,
		logger:   logger,
		entries:  make(map[string]*cronEntry),
	}
}

// Start begins the scheduler loop. It reloads triggers from the store
// whenever they change and fires matching entries every minute.
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

	triggers := s.store.ListTriggers()
	newEntries := make(map[string]*cronEntry, len(triggers))

	for _, t := range triggers {
		if t.Type != store.TriggerTypeCron || !t.Enabled || t.Cron == nil {
			continue
		}

		sched, err := parseCron(t.Cron.Schedule)
		if err != nil {
			s.logger.Warn("Invalid cron schedule, skipping", "trigger", t.Name, "schedule", t.Cron.Schedule, "error", err)
			continue
		}

		if existing, ok := s.entries[t.ID]; ok {
			existing.trigger = t
			existing.schedule = sched
			newEntries[t.ID] = existing
		} else {
			newEntries[t.ID] = &cronEntry{
				trigger:  t,
				schedule: sched,
				next:     sched.Next(time.Now()),
			}
		}
	}

	s.entries = newEntries
	s.logger.Debug("Scheduler reloaded", "cronTriggers", len(newEntries))
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
		go func(t store.Trigger) {
			s.logger.Info("Cron trigger firing", "trigger", t.Name, "id", t.ID)
			result, err := s.executor.RunTrigger(ctx, t, "", "")
			if err != nil {
				s.logger.Error("Cron trigger failed", "trigger", t.Name, "error", err)
				return
			}
			s.logger.Info("Cron trigger completed", "trigger", t.Name, "responseLen", len(result))
		}(entry.trigger)
	}
}
