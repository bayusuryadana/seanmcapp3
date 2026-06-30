package service

import (
	"log"
	"sync/atomic"
)

type runGuard struct {
	running atomic.Bool
}

func (g *runGuard) run(name string, fn func()) {
	if !g.running.CompareAndSwap(false, true) {
		log.Printf("[INFO] %s already in progress, skipping", name)
		return
	}
	defer g.running.Store(false)
	fn()
}
