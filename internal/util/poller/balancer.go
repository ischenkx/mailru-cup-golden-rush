package poller

import (
	"sync"
	"time"
)

type Balancer struct {
	p Poller
	maxTimeout time.Duration
	counter int
	breakPoint int
	mu sync.RWMutex
}

func (b *Balancer) Get() Poller {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.p
}

func (b *Balancer) Register(to time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if to > b.maxTimeout {
		b.maxTimeout = to
	}
	b.counter++

	if b.counter >= b.breakPoint {
		b.counter = 0
		b.p.timeOut = 3 * b.maxTimeout / 2
	}
}

func NewBalancer(p Poller, bp int) *Balancer {
	b := &Balancer{
		p:          p,
		maxTimeout: 0,
		counter:    0,
		breakPoint: bp,
	}
	return b
}

