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
