package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",  
		Password: "",                
		DB:       0,                
	})
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}