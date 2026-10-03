package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	c *cron.Cron
}

func New(loc *time.Location) *Scheduler {
	return &Scheduler{c: cron.New(
		cron.WithLocation(loc),
		cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DefaultLogger)),
	)}
}

func (s *Scheduler) AddJob(spec string, timeout time.Duration, fn func(ctx context.Context)) error {
	_, err := s.c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		fn(ctx)
	})
	return err
}

func (s *Scheduler) Start() {
	s.c.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.c.Stop()
}
