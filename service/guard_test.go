package service

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunGuard(t *testing.T) {
	var g runGuard
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	var secondRan atomic.Bool

	go func() {
		g.run("job", func() {
			close(started)
			<-release
		})
		close(done)
	}()

	<-started
	g.run("job", func() { secondRan.Store(true) })
	assert.False(t, secondRan.Load(), "concurrent run should be skipped")

	close(release)
	<-done

	g.run("job", func() { secondRan.Store(true) })
	assert.True(t, secondRan.Load(), "run after release should execute")
}
