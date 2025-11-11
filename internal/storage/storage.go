package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("key not found")
)

// Storage is the interface for all storage backends
type Storage interface {
	// Event storage
	SetEvent(ctx context.Context, eventID string, data []byte) error
	GetEvent(ctx context.Context, eventID string) ([]byte, error)
	ListEventIDs(ctx context.Context) ([]string, error)

	// Feed storage (domain sets)
	AddDomain(ctx context.Context, feed, domain string) error
	GetDomains(ctx context.Context, feed string) ([]string, error)
	RemoveDomain(ctx context.Context, feed, domain string) error
	ListFeeds(ctx context.Context) ([]string, error)

	// Feed metadata
	SetFeedMeta(ctx context.Context, feed, key, value string) error
	GetFeedMeta(ctx context.Context, feed, key string) (string, error)

	// Generic key-value
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error

	// STIX object storage
	SetSTIXObject(ctx context.Context, stixID string, data []byte) error
	GetSTIXObject(ctx context.Context, stixID string) ([]byte, error)
	ListSTIXObjects(ctx context.Context) ([]string, error)
	DeleteSTIXObject(ctx context.Context, stixID string) error

	// STIX ID mapping (domain -> stix_id)
	SetDomainStixID(ctx context.Context, domain, stixID string) error
	GetDomainStixID(ctx context.Context, domain string) (string, error)

	// Close the storage connection
	Close() error
}
