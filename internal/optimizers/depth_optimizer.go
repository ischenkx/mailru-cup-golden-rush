package optimizers

import (
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"sync"
)

type depthInfo struct {
	time float64
	treasures float64
	counter float64
}

func (info *depthInfo) register(time, treasures int64) {
	info.treasures = info.counter / (info.counter+1) * info.treasures + float64(treasures) / (info.counter+1)
	info.time = (info.counter / (info.counter+1)) * info.time + float64(time) / (info.counter+1)
	info.counter += 1
}

type DepthOptimizer struct {
	k, g int64
	nextCounter int64
	registerCounter int64
	maxDepth int64
	currentBest int64
	counters []depthInfo
	mu sync.RWMutex
}

func (d *DepthOptimizer) Next() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextCounter++
	if d.nextCounter % d.k == 0 {
		d.nextCounter = 0
		return d.maxDepth
	}
	return d.currentBest
}

func (d *DepthOptimizer) recalculate() {
	best := float64(d.currentBest)
	bestRatio := 0.
	currentTime := 0.
	currentTreasures := 0.
	for i, info := range d.counters {
		currentTime += info.time
		currentTreasures += info.treasures
		ratio := currentTreasures/currentTime
		if ratio > bestRatio {
			best = float64(i+1)
			bestRatio = ratio
		}
	}
	d.currentBest = int64(best)
}

func (d *DepthOptimizer) Register(depth, treasures, time int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	info := &d.counters[depth-1]
	info.register(time, treasures)
	d.registerCounter++
	if d.registerCounter % d.g == 0 {
		d.recalculate()
	}
}

func (d *DepthOptimizer) Best() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentBest
}

func NewDepthOptimizer(cfg config.Config) *DepthOptimizer {
	return &DepthOptimizer{
		k:               int64(cfg.App.World.DepthOptimizer.K),
		g:               int64(cfg.App.World.DepthOptimizer.G),
		nextCounter:     0,
		registerCounter: 0,
		maxDepth:        int64(cfg.App.World.Depth),
		currentBest:     int64(cfg.App.World.Depth),
		counters:        make([]depthInfo, cfg.App.World.Depth),
		mu:              sync.RWMutex{},
	}
}