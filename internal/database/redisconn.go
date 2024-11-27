package database

import (
	"Games/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

var (
	RedisClient *redis.Client
	ctx         context.Context
)

func ConnectRedis(config *config.Config) {
	ctx = context.TODO()
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.RedisHost + ":" + config.RedisPort,
	})

	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	err := RedisClient.Set(ctx, "test", "test", 0).Err()
	if err != nil {
		panic(err)
	}

	log.Println("✅ Redis client connected successfully...")
}
