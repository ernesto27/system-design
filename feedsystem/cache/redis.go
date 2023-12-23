package cache

import (
	"github.com/go-redis/redis"
)

type Redis struct {
	redis *redis.Client
}

func NewRedis(host string, port string) *Redis {
	rc := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
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

func (r *Redis) ListSet(key, value string) error {
	err := r.redis.LPush(key, value).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) ListGet(key string) ([]string, error) {
	posts, err := r.redis.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return posts, nil

}
