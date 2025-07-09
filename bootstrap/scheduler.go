package bootstrap

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func InitScheduler(mainServices MainServices) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	c := cron.New(cron.WithSeconds(), cron.WithLocation(loc))

	schedulers := []*Scheduler{
		{Task: mainServices.WarmupDBService, CronExpr: "0 * * * * *", Repeat: false},
		{Task: mainServices.WarmupDBService, CronExpr: "*/5 * * * * *", Repeat: false},
		{Task: mainServices.BirthdayService, CronExpr: "0 0 8 * * *", Repeat: true},
	}

	for _, s := range schedulers {
		_, err := s.Schedule(c)
		if err != nil {
			log.Fatalf("Failed to schedule task: %v", err)
		}
	}

	fmt.Println("Running scheduled jobs...")
	c.Start()
}

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
