package storage

import (
	"context"
	"sync"
)

// MemoryStorage is an in-memory implementation of Storage
type MemoryStorage struct {
	mu      sync.RWMutex
	events  map[string][]byte
	feeds   map[string]map[string]bool // feed -> domain set
	kvStore map[string][]byte
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events:  make(map[string][]byte),
		feeds:   make(map[string]map[string]bool),
		kvStore: make(map[string][]byte),
	}
}

func (m *MemoryStorage) SetEvent(ctx context.Context, eventID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[eventID] = data
	return nil
}

func (m *MemoryStorage) GetEvent(ctx context.Context, eventID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.events[eventID]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *MemoryStorage) ListEventIDs(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.events))
	for id := range m.events {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MemoryStorage) AddDomain(ctx context.Context, feed, domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.feeds[feed] == nil {
		m.feeds[feed] = make(map[string]bool)
	}
	m.feeds[feed][domain] = true
	return nil
}

func (m *MemoryStorage) GetDomains(ctx context.Context, feed string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	domainSet, ok := m.feeds[feed]
	if !ok {
		return []string{}, nil
	}
	domains := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}
	return domains, nil
}

func (m *MemoryStorage) RemoveDomain(ctx context.Context, feed, domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.feeds[feed] != nil {
		delete(m.feeds[feed], domain)
	}
	return nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kvStore[key] = value
	return nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.kvStore[key]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.kvStore, key)
	return nil
}

func (m *MemoryStorage) Close() error {
	return nil
}
