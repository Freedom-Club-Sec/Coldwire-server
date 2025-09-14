package redis

import (
    "github.com/redis/go-redis/v9"
    "context"

    "github.com/Freedom-Club-Sec/Coldwire-server/internal/storage"
)

type RedisStorage struct {
    client *redis.Client
}

func New(addr string) *RedisStorage {
    rdb := redis.NewClient(&redis.Options{Addr: addr})
    return &RedisStorage{client: rdb}
}

func (r *RedisStorage) SaveChallenge(challenge string, id string, publicKey string) error {
    data := []string{id, publicKey}
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return err
    }

    jsonString = string(jsonBytes)

    key := "challenges:" + challenge

    return r.client.Set(context.Background(), key, jsonString, 0).Err()
}

func (r *RedisStorage) GetChallengeData(challenge string) (string, string, error) {
    val, err := r.client.Get(context.Background(), challenge).Result()
    if err != nil {
        return "", "", err
    }

    key := "challenges:" + challenge
    
    r.DeleteKey(key)

    var data []string
    err = json.Unmarshal(val, &data)
    if err != nil {
        return "", "", err
    }

    return data[0], data[1], nil
}

func (r *RedisStorage) DeleteKey(key string) error {
    return r.client.Del(context.Background(), key).Err()
}

