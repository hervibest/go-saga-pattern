package config

import (
	"context"
	"go-saga-pattern/commoner/utils"
	"log"

	"github.com/redis/go-redis/v9"
)

// TODO LOCAL VARIABLE (DEVEL, PROD)
func NewRedisClient() *redis.Client {
	host := utils.GetEnv("REDIS_HOST")
	port := utils.GetEnv("REDIS_PORT")
	address := host + ":" + port

	password := ""

	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	ctx := context.TODO()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	log.Println("✅ Redis client connected successfully...")

	return rdb
}
