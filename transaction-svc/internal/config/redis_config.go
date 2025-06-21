package config

import (
	"context"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TODO LOCAL VARIABLE (DEVEL, PROD)
func NewRedisClient(log logs.Log) *redis.Client {
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
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	log.Info("Connected to Redis", zap.String("address", address))

	return rdb
}
