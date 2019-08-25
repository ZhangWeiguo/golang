package client

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisClient struct {
	Addr   string
	Pass   string
	DB     int
	client *redis.Client
}

func (rc *RedisClient) Init() error {
	rc.client = redis.NewClient(&redis.Options{
		Addr:     rc.Addr,
		Password: rc.Pass,
		DB:       rc.DB,
	})
	_, err := rc.client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	result := rc.client.Set(key, value, expiration)
	return result.Err()
}

func (rc *RedisClient) Get(key string) {
	result := rc.client.Get(key)
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
}
