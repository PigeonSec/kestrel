package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ValkeyStorage is a Valkey/Redis implementation of Storage
type ValkeyStorage struct {
	client *redis.Client
}

// NewValkeyStorage creates a new Valkey/Redis storage
func NewValkeyStorage(addr string) (*ValkeyStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &ValkeyStorage{
		client: client,
	}, nil
}

func (v *ValkeyStorage) SetEvent(ctx context.Context, eventID string, data []byte) error {
	pipe := v.client.Pipeline()
	pipe.Set(ctx, "misp:event:"+eventID, data, 0)
	pipe.LPush(ctx, "misp:events", eventID)
	_, err := pipe.Exec(ctx)
	return err
}

func (v *ValkeyStorage) GetEvent(ctx context.Context, eventID string) ([]byte, error) {
	data, err := v.client.Get(ctx, "misp:event:"+eventID).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return data, err
}

func (v *ValkeyStorage) ListEventIDs(ctx context.Context) ([]string, error) {
	return v.client.LRange(ctx, "misp:events", 0, -1).Result()
}

func (v *ValkeyStorage) AddDomain(ctx context.Context, feed, domain string) error {
	return v.client.SAdd(ctx, "misp:feed:"+feed, domain).Err()
}

func (v *ValkeyStorage) GetDomains(ctx context.Context, feed string) ([]string, error) {
	return v.client.SMembers(ctx, "misp:feed:"+feed).Result()
}

func (v *ValkeyStorage) RemoveDomain(ctx context.Context, feed, domain string) error {
	return v.client.SRem(ctx, "misp:feed:"+feed, domain).Err()
}

func (v *ValkeyStorage) ListFeeds(ctx context.Context) ([]string, error) {
	keys, err := v.client.Keys(ctx, "misp:feed:*").Result()
	if err != nil {
		return nil, err
	}

	feeds := make([]string, 0, len(keys))
	for _, key := range keys {
		// Skip metadata keys
		if len(key) > 14 && key[:14] == "misp:feed:meta:" {
			continue
		}
		// Extract feed name from key (remove "misp:feed:" prefix)
		if len(key) > 10 {
			feedName := key[10:]
			feeds = append(feeds, feedName)
		}
	}
	return feeds, nil
}

func (v *ValkeyStorage) SetFeedMeta(ctx context.Context, feed, key, value string) error {
	return v.client.HSet(ctx, "misp:feed:meta:"+feed, key, value).Err()
}

func (v *ValkeyStorage) GetFeedMeta(ctx context.Context, feed, key string) (string, error) {
	value, err := v.client.HGet(ctx, "misp:feed:meta:"+feed, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return value, err
}

func (v *ValkeyStorage) Set(ctx context.Context, key string, value []byte) error {
	return v.client.Set(ctx, key, value, 0).Err()
}

func (v *ValkeyStorage) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := v.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return data, err
}

func (v *ValkeyStorage) Delete(ctx context.Context, key string) error {
	return v.client.Del(ctx, key).Err()
}

func (v *ValkeyStorage) Close() error {
	return v.client.Close()
}

// STIX object storage methods

func (v *ValkeyStorage) SetSTIXObject(ctx context.Context, stixID string, data []byte) error {
	pipe := v.client.Pipeline()
	pipe.Set(ctx, "stix:object:"+stixID, data, 0)
	pipe.LPush(ctx, "stix:objects", stixID)
	_, err := pipe.Exec(ctx)
	return err
}

func (v *ValkeyStorage) GetSTIXObject(ctx context.Context, stixID string) ([]byte, error) {
	data, err := v.client.Get(ctx, "stix:object:"+stixID).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return data, err
}

func (v *ValkeyStorage) ListSTIXObjects(ctx context.Context) ([]string, error) {
	return v.client.LRange(ctx, "stix:objects", 0, -1).Result()
}

func (v *ValkeyStorage) DeleteSTIXObject(ctx context.Context, stixID string) error {
	pipe := v.client.Pipeline()
	pipe.Del(ctx, "stix:object:"+stixID)
	pipe.LRem(ctx, "stix:objects", 0, stixID)
	_, err := pipe.Exec(ctx)
	return err
}

// STIX ID mapping methods

func (v *ValkeyStorage) SetDomainStixID(ctx context.Context, domain, stixID string) error {
	return v.client.Set(ctx, "stix:domain:"+domain, stixID, 0).Err()
}

func (v *ValkeyStorage) GetDomainStixID(ctx context.Context, domain string) (string, error) {
	stixID, err := v.client.Get(ctx, "stix:domain:"+domain).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return stixID, err
}
