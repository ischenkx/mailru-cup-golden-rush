package poller

import (
	"errors"
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"time"
)

var MaxIterationsReachedErr = errors.New("max iterations reached")
var DeadlineReachedErr = errors.New("deadline reached")
type Poller struct {
	timeout time.Duration
	interval time.Duration
	maxIterations int
}

func (p *Poller) Do(f func(deadline time.Time) (interface{}, error, bool)) (interface{}, error) {
	i := 0
	for {
		if i >= p.maxIterations && p.maxIterations > 0 {
			return nil, MaxIterationsReachedErr
		}
		i++
		now := time.Now()
		deadline := now.Add(p.timeout)
		data, err, ok := f(deadline)
		if ok {
			return data, err
		}

		if p.interval > 0 {
			time.Sleep(p.interval)
		}
	}
}

func (p *Poller) DoDeadline(deadline time.Time, f func(deadline time.Time) (interface{}, error, bool)) (interface{}, error) {
	i := 0
	for {
		if time.Now().After(deadline) {
			return nil, DeadlineReachedErr
		}
		if i >= p.maxIterations && p.maxIterations > 0 {
			return nil, MaxIterationsReachedErr
		}
		i++
		data, err, ok := f(deadline)
		if ok {
			return data, err
		}
		if p.interval > 0 {
			time.Sleep(p.interval)
		}
	}
}

func FromConfig(cfg config.PollerConfig) *Poller {
	return &Poller{
		timeout:       cfg.TimeOut.Parse(),
		interval:      cfg.Interval.Parse(),
		maxIterations: cfg.MaxIters,
	}
}

func NewPoller(timeout time.Duration, maxIterations int, interval time.Duration) *Poller {
	return &Poller{
		timeout:       timeout,
		maxIterations: maxIterations,
		interval: interval,
	}
}