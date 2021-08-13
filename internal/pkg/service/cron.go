package service

import (
	"github.com/go-co-op/gocron"
	"time"
)

// BuildCron /* *gocron.Scheduler
func BuildCron() (*gocron.Scheduler, error) {
	return gocron.NewScheduler(time.UTC), nil
}
