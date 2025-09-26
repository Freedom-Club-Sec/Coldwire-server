package redis

import (
	"bytes"
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

	result, err := s.client.LRange(ctx, userId, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var allData []byte
	for _, value := range result {
		allData = append(allData, []byte(value)...)
	}

	return allData, nil
}

func (s *RedisStorage) DeleteAck(userId string, acks [][]byte) error {
	ctx := context.Background()
	values, err := s.client.LRange(ctx, userId, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, v := range values {
		data := []byte(v)

		for _, ackId := range acks {
			if bytes.HasPrefix(data, ackId) {
				if err := s.client.LRem(ctx, userId, 0, v).Err(); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}
func (s *RedisStorage) InsertData(dataBlob []byte, ackId []byte, recipientId string) error {
	dataBlob = append(ackId, dataBlob...)
	return s.client.RPush(context.Background(), recipientId, dataBlob).Err()
}

func (s *RedisStorage) ExitCleanup() error {
	return nil
}
