package license

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Handle struct {
	desc   *int64
	m      *Manager
	id     int64
	active bool
}

func (h Handle) Active() bool {
	return h.active
}

func (h *Handle) Close() {
	h.active = false
	if atomic.AddInt64(h.desc, -1) == 0 {
		h.m.deleteLicense(h.id)
	}
}

func (h Handle) ID() int64 {
	return h.id
}

type AddHandle struct {
	done bool
	m *Manager
}

func (h *AddHandle) Ok(id, digs int64) error {
	if h.done {
		return errors.New("already done")
	}
	h.done = true

	return h.m.add(id, digs)
}

func (h *AddHandle) Fail() {
	if h.done {
		return
	}
	h.done = true
	h.m.mu.Lock()
	h.m.licenseCounter--
	h.m.addCond.Signal()
	h.m.mu.Unlock()
}


type Manager struct {
	maxLicenses      int
	licenseCounter int
	addCond, getCond *sync.Cond
	licenses         map[int64]license
	licensesInUse    map[int64]license
	deletedLicenses  map[int64]struct{}
	mu               sync.RWMutex
}

func (m *Manager) deleteLicense(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.licensesInUse[id]; ok {
		l.Close()
		delete(m.licenses, id)
		delete(m.licensesInUse, id)
		m.deletedLicenses[id] = struct{}{}
		m.licenseCounter--
		m.addCond.Signal()
	}
}

func (m *Manager) RequestAdd() AddHandle {
	m.mu.Lock()
	defer m.mu.Unlock()
	for m.licenseCounter >= m.maxLicenses {
		m.addCond.Wait()
	}
	m.licenseCounter++
	return AddHandle{
		done: false,
		m:    m,
	}
}

func (m *Manager) add(id, digs int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if digs < 1 {
		return errors.New("cannot use 'digs' that are less than 1")
	}

	if _, ok := m.licenses[id]; ok {
		return errors.New("license already registered")
	}

	if _, ok := m.licensesInUse[id]; ok {
		return errors.New("license already registered")
	}

	if _, ok := m.deletedLicenses[id]; ok {
		return errors.New("license already registered")
	}
	m.licenses[id] = newLicense(id, digs)
	m.getCond.Broadcast()
	return nil
}

func (m *Manager) Get() Handle {
	m.mu.Lock()
	for len(m.licenses) == 0 {
		m.getCond.Wait()
	}
	for _, l := range m.licenses {
		l.digs--
		h := Handle{
			desc:   l.desc,
			m:      m,
			id:     l.id,
			active: true,
		}

		if l.digs <= 0 {
			m.licensesInUse[l.id] = l
			delete(m.licenses, l.id)
		} else {
			m.licenses[l.id] = l
		}
		m.mu.Unlock()
		return h
	}
	m.mu.Unlock()
	return m.Get()
}

func NewManager(maxLicenses int) *Manager {

	m := &Manager{
		maxLicenses:     maxLicenses,
		licenseCounter:  0,
		licenses:        map[int64]license{},
		licensesInUse:   map[int64]license{},
		deletedLicenses: map[int64]struct{}{},
		mu:              sync.RWMutex{},
	}

	m.getCond = sync.NewCond(&m.mu)
	m.addCond = sync.NewCond(&m.mu)

	return m
}
