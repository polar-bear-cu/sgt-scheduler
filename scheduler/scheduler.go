package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	c *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{c: cron.New()}
}

func (s *Scheduler) AddJob(spec string, fn func(ctx context.Context)) error {
	_, err := s.c.AddFunc(spec, func() {
		fn(context.Background())
	})
	return err
}

func (s *Scheduler) Start() {
	s.c.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.c.Stop()
}
