package coin

import "sync"

type Manager struct {
	mu sync.RWMutex
	coins []uint32
}

func (w *Manager) Pop(n int) []uint32 {
	w.mu.Lock()
	defer w.mu.Unlock()

	l := len(w.coins)

	if l < n {
		n = len(w.coins)
	}

	wallet := make([]uint32, n)

	copy(wallet, w.coins[l-n:])
	w.coins = w.coins[:l-n]

	return wallet
}

func (w *Manager) Add(coins ...uint32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.coins = append(w.coins, coins...)
}

func (w *Manager) Amount() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := len(w.coins)
	return int64(s)
}

func NewManager() *Manager {
	return &Manager{
		mu:    sync.RWMutex{},
		coins: make([]uint32, 0, 1000000),
	}
}