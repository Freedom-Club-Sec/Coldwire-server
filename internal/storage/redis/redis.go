package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func New(addr string, port string, password string, db int) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr + ":" + port,
		Password: password,
		DB:       db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisStorage{client: rdb}, nil
}

// / Implements DataStorage interface
func (s *RedisStorage) GetLatestData(userId string) ([]byte, error) {
	ctx := context.Background()

	// Lua script: get current items, remove them from the list
	luaScript := `
        local len = redis.call('LLEN', KEYS[1])
        if len == 0 then
            return {}
        end
        local items = redis.call('LRANGE', KEYS[1], 0, len - 1)
        redis.call('LTRIM', KEYS[1], len, -1)
        return items
    `

	// Execute Lua script
	result, err := s.client.Eval(ctx, luaScript, []string{userId}).Result()
	if err != nil {
		return nil, err
	}

	// Convert []interface{} to []byte
	var allData []byte
	if items, ok := result.([]interface{}); ok {
		for _, item := range items {
			if b, ok := item.(string); ok {
				allData = append(allData, []byte(b)...)
			}
		}
	}

	return allData, nil
}

func (s *RedisStorage) InsertData(dataBlob []byte, recipientId string) error {
	return s.client.RPush(context.Background(), recipientId, dataBlob).Err()
}

func (s *RedisStorage) ExitCleanup() error {
	return nil
}
