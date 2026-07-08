package bootstrap

import (
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTask struct{ runs int }

func (f *fakeTask) Run() { f.runs++ }

func TestScheduleInvalidExpr(t *testing.T) {
	c := cron.New(cron.WithSeconds())
	s := &Scheduler{Task: &fakeTask{}, CronExpr: "not-a-cron", Repeat: true}
	_, err := s.Schedule(c)
	assert.Error(t, err)
}

func TestScheduleRepeat(t *testing.T) {
	c := cron.New(cron.WithSeconds())
	task := &fakeTask{}
	s := &Scheduler{Task: task, CronExpr: "0 0 9 * * *", Repeat: true}

	id, err := s.Schedule(c)
	require.NoError(t, err)

	// Trigger the scheduled job manually (cron is not started).
	c.Entry(id).Job.Run()

	assert.Equal(t, 1, task.runs)
	assert.Len(t, c.Entries(), 1, "repeat task should remain scheduled after running")
}

func TestScheduleOnce(t *testing.T) {
	c := cron.New(cron.WithSeconds())
	task := &fakeTask{}
	s := &Scheduler{Task: task, CronExpr: "0 0 9 * * *", Repeat: false}

	id, err := s.Schedule(c)
	require.NoError(t, err)

	c.Entry(id).Job.Run()

	assert.Equal(t, 1, task.runs)
	assert.Empty(t, c.Entries(), "one-shot task should remove itself after running")
}

