package cache

import (
	"github.com/go-redis/redis"
)

type Redis struct {
	redis *redis.Client
}

func NewRedis(host string) *Redis {
	rc := redis.NewClient(&redis.Options{
		Addr:     host + ":6379",
		Password: "",
		DB:       0,
	})

	return &Redis{
		redis: rc,
	}
}

func (r *Redis) Set(key string, value interface{}) error {
	return r.redis.Set(key, value, 0).Err()
}

func (r *Redis) Get(key string) (string, error) {
	return r.redis.Get(key).Result()
}
