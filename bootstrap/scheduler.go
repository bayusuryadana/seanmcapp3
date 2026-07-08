package bootstrap

import (
	"github.com/robfig/cron/v3"
)


type ScheduledTask interface {
	Run()
}

type Scheduler struct {
	Task     ScheduledTask
	CronExpr string
	Repeat   bool
}

func (s *Scheduler) Schedule(cronEngine *cron.Cron) (cron.EntryID, error) {
	var entryID cron.EntryID
	var err error

	entryID, err = cronEngine.AddFunc(s.CronExpr, func() {
		s.Task.Run()
		if !s.Repeat {
			cronEngine.Remove(entryID)
		}
	})

	return entryID, err
}
