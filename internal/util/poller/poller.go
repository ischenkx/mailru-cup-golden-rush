package poller

import (
	"context"
	"fmt"
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"time"
)

type Poller struct {
	timeOut    time.Duration
	interval   time.Duration
	maxIters   int
	ctx        context.Context
	errHandler func(error) (error, bool)
}

type PollResult struct {
	data        interface{}
	err         error
	iters		int
	deadlinesHit int
	okRequestTime time.Duration
	elapsedTime time.Duration
}

func (r PollResult) DeadlinesHit() int {
	return r.deadlinesHit
}

func (r PollResult) Data() interface{} {
	return r.data
}

func (r PollResult) Error() error {
	return r.err
}

func (r PollResult) ElapsedTime() time.Duration {
	return r.elapsedTime
}

func (r PollResult) OkRequestTime() time.Duration {
	return r.okRequestTime
}

func (r PollResult) Iters() int {
	return r.iters
}

func (p *Poller) validate() {
	if p.timeOut == 0 {
		p.timeOut = time.Second * 60
	}
	if p.errHandler == nil {
		p.errHandler = func(err error) (error, bool) {
			return err, err == nil
		}
	}
	if p.maxIters == 0 {
		p.maxIters = 10
	}
	if p.ctx == nil {
		p.ctx = context.Background()
	}
}

func (p Poller) WithTimeOut(to time.Duration) Poller {
	p.timeOut = to
	return p
}

func (p Poller) WithMaxIters(i int) Poller {
	p.maxIters = i
	return p
}

func (p Poller) WithInterval(interval time.Duration) Poller {
	p.interval = interval
	return p
}

func (p Poller) WithContext(ctx context.Context) Poller {
	p.ctx = ctx
	return p
}

func (p Poller) HandleErrors(f func(error) (error, bool)) Poller {
	p.errHandler = f
	return p
}

func (p Poller) Do(f func(t time.Duration) (interface{}, error)) (result PollResult) {
	p.validate()
	i := 0
	startTime := time.Now()
	for {
		if p.maxIters > 0 && i > p.maxIters {
			result.err = fmt.Errorf("poll failure - max iters (%d)", p.maxIters)
			break
		}
		lastAttemptStartTime := time.Now()

		res, err := f(p.timeOut)

		//if errors.Is(err, context.DeadlineExceeded) {
		//	result.deadlinesHit++
		//}

		if e, ok := p.errHandler(err); ok {
			result.err = e
			result.data = res
			result.okRequestTime = time.Since(lastAttemptStartTime)
			break
		}
		i++

		if p.interval > 0 {
			time.Sleep(p.interval)
		}
	}
	result.iters = i
	result.elapsedTime = time.Since(startTime)
	return
}

func FromConfig(cfg config.PollerConfig) Poller {
	d, p := cfg.TimeOut.Parse(), cfg.Interval.Parse()
	fmt.Println(cfg.TimeOut, d, p)
	return Poller{
		timeOut:    cfg.TimeOut.Parse(),
		interval:   cfg.Interval.Parse(),
		maxIters:   cfg.MaxIters,
	}
}

func New(ctx context.Context) Poller {
	return Poller{ctx: ctx}
}
