package storage

import (
	"context"
	"sync"
)

// MemoryStorage is an in-memory implementation of Storage
type MemoryStorage struct {
	mu            sync.RWMutex
	events        map[string][]byte
	feeds         map[string]map[string]bool   // feed -> domain set
	feedMeta      map[string]map[string]string // feed -> metadata map
	kvStore       map[string][]byte
	stixObjects   map[string][]byte            // stix_id -> stix object JSON
	domainStixIDs map[string]string            // domain -> stix_id mapping
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events:        make(map[string][]byte),
		feeds:         make(map[string]map[string]bool),
		feedMeta:      make(map[string]map[string]string),
		kvStore:       make(map[string][]byte),
		stixObjects:   make(map[string][]byte),
		domainStixIDs: make(map[string]string),
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

func (m *MemoryStorage) ListFeeds(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	feeds := make([]string, 0, len(m.feeds))
	for feed := range m.feeds {
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (m *MemoryStorage) SetFeedMeta(ctx context.Context, feed, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.feedMeta[feed] == nil {
		m.feedMeta[feed] = make(map[string]string)
	}
	m.feedMeta[feed][key] = value
	return nil
}

func (m *MemoryStorage) GetFeedMeta(ctx context.Context, feed, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.feedMeta[feed] == nil {
		return "", ErrNotFound
	}
	value, ok := m.feedMeta[feed][key]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
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

// STIX object storage methods

func (m *MemoryStorage) SetSTIXObject(ctx context.Context, stixID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stixObjects[stixID] = data
	return nil
}

func (m *MemoryStorage) GetSTIXObject(ctx context.Context, stixID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.stixObjects[stixID]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *MemoryStorage) ListSTIXObjects(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.stixObjects))
	for id := range m.stixObjects {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MemoryStorage) DeleteSTIXObject(ctx context.Context, stixID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.stixObjects, stixID)
	return nil
}

// STIX ID mapping methods

func (m *MemoryStorage) SetDomainStixID(ctx context.Context, domain, stixID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.domainStixIDs[domain] = stixID
	return nil
}

func (m *MemoryStorage) GetDomainStixID(ctx context.Context, domain string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stixID, ok := m.domainStixIDs[domain]
	if !ok {
		return "", ErrNotFound
	}
	return stixID, nil
}
