package bootstrap

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func InitScheduler(mainServices MainServices) *cron.Cron {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(loc),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	schedulers := []*Scheduler{
		{Task: mainServices.NewsService, CronExpr: "0 0 9 * * *", Repeat: true},
		{Task: mainServices.StockService, CronExpr: "0 0 19 * * *", Repeat: true},
		{Task: mainServices.InstagramService, CronExpr: "0 0 10 * * *", Repeat: true},
	}

	for _, s := range schedulers {
		_, err := s.Schedule(c)
		if err != nil {
			log.Fatalf("Failed to schedule task: %v", err)
		}
	}

	log.Println("Running scheduled jobs...")
	c.Start()
	return c
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
