package mertics

import (
	"sync"
)

type avg struct {
	cnt, val float64
}

func (a *avg) add(f float64) {
	a.val = (a.val*a.cnt + f) / (a.cnt + 1)
	a.cnt++
}

type Snapshot struct {
	Counters map[string]float64 `json:"counters"`
	Max      map[string]float64 `json:"max"`
	Average  map[string]float64 `json:"average"`
}

type Metrics struct {
	enabled  bool
	counters map[string]float64
	max      map[string]float64
	average  map[string]avg
	mu       sync.RWMutex
}

func (m *Metrics) AddAverage(key string, v float64) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := m.average[key]
	entry.add(v)
	m.average[key] = entry
}

func (m *Metrics) AddCounter(key string, v float64) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[key] += v
}

func (m *Metrics) AddMax(key string, v float64) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	val, ok := m.max[key]

	if !ok || val < v {
		m.max[key] = v
	}
}

func (m *Metrics) IncCounter(key string) {
	if !m.enabled {
		return
	}
	m.AddCounter(key, 1)
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := Snapshot{
		Counters: map[string]float64{},
		Max:      map[string]float64{},
		Average:  map[string]float64{},
	}

	for key, val := range m.max {
		s.Max[key] = val
	}

	for key, val := range m.average {
		s.Average[key] = val.val
	}
	for key, val := range m.counters {
		s.Counters[key] = val
	}
	return s
}

func New(enabled bool) *Metrics {
	return &Metrics{
		enabled:  enabled,
		counters: map[string]float64{},
		max:      map[string]float64{},
		average:  map[string]avg{},
		mu:       sync.RWMutex{},
	}
}
