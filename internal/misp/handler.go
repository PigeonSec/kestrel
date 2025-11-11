package misp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pigeonsec/kestrel/internal/storage"
)

// Handler manages MISP events with in-memory caching
type Handler struct {
	storage      storage.Storage
	eventCache   sync.Map       // id -> []byte
	manifestJSON atomic.Value   // []byte
}

// NewHandler creates a new MISP handler
func NewHandler(storage storage.Storage) *Handler {
	h := &Handler{
		storage: storage,
	}

	// Initialize manifest
	h.updateManifest()

	return h
}

// CreateEvent creates a new MISP event
func (h *Handler) CreateEvent(ctx context.Context, eventID, domain, category, comment string) error {
	event := EventWrapper{
		Event: Event{
			Info:          "IOC from Kestrel API",
			ThreatLevelID: "3",
			Analysis:      "0",
			Distribution:  "0",
			Attribute: []Attribute{
				{
					Type:      "domain",
					Category:  category,
					Value:     domain,
					ToIDs:     true,
					Comment:   comment,
					Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
				},
			},
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Store in cache
	h.eventCache.Store(eventID, data)

	// Store in storage backend
	if err := h.storage.SetEvent(ctx, eventID, data); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Update manifest
	h.updateManifest()

	return nil
}

// GetEvent retrieves a MISP event by ID
func (h *Handler) GetEvent(ctx context.Context, eventID string) ([]byte, error) {
	// Check cache first
	if data, ok := h.eventCache.Load(eventID); ok {
		return data.([]byte), nil
	}

	// Fallback to storage
	data, err := h.storage.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Cache for next time
	h.eventCache.Store(eventID, data)

	return data, nil
}

// GetManifest returns the MISP manifest
func (h *Handler) GetManifest() []byte {
	v := h.manifestJSON.Load()
	if v == nil {
		return []byte("{}")
	}
	return v.([]byte)
}

// updateManifest updates the cached manifest
func (h *Handler) updateManifest() {
	manifest := make(map[string]ManifestEntry)

	h.eventCache.Range(func(key, value interface{}) bool {
		id := key.(string)
		manifest[id+".json"] = ManifestEntry{UUID: id}
		return true
	})

	data, _ := json.Marshal(manifest)
	h.manifestJSON.Store(data)
}

// LoadEventsFromStorage loads all events from storage into cache
func (h *Handler) LoadEventsFromStorage(ctx context.Context) error {
	eventIDs, err := h.storage.ListEventIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range eventIDs {
		data, err := h.storage.GetEvent(ctx, id)
		if err != nil {
			continue
		}
		h.eventCache.Store(id, data)
	}

	h.updateManifest()
	return nil
}

// GetAllEvents returns all MISP events as a JSON array
func (h *Handler) GetAllEvents(ctx context.Context) []byte {
	var events []json.RawMessage

	h.eventCache.Range(func(key, value interface{}) bool {
		data := value.([]byte)
		events = append(events, json.RawMessage(data))
		return true
	})

	// If no cached events, try loading from storage
	if len(events) == 0 {
		eventIDs, err := h.storage.ListEventIDs(ctx)
		if err == nil {
			for _, id := range eventIDs {
				data, err := h.storage.GetEvent(ctx, id)
				if err == nil {
					events = append(events, json.RawMessage(data))
				}
			}
		}
	}

	result, _ := json.Marshal(events)
	return result
}
