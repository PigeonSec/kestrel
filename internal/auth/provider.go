package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// KeyProvider is an interface for fetching API keys from external sources
type KeyProvider interface {
	FetchKeys(ctx context.Context) ([]*Account, error)
}

// HTTPKeyProvider fetches keys from an HTTP endpoint
type HTTPKeyProvider struct {
	url    string
	client *http.Client
}

// NewHTTPKeyProvider creates a new HTTP key provider
func NewHTTPKeyProvider(url string) *HTTPKeyProvider {
	return &HTTPKeyProvider{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchKeys fetches keys from the HTTP endpoint
func (p *HTTPKeyProvider) FetchKeys(ctx context.Context) ([]*Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var accounts []*Account
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

// StartProviderSync starts periodic synchronization from external provider
func StartProviderSync(provider KeyProvider, store *KeyStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			accounts, err := provider.FetchKeys(ctx)
			cancel()

			if err != nil {
				log.Printf("Failed to fetch keys from provider: %v", err)
				continue
			}

			if err := store.ReplaceAll(accounts); err != nil {
				log.Printf("Failed to update key store: %v", err)
				continue
			}

			log.Printf("Synced %d accounts from external provider", len(accounts))
		}
	}()
}
