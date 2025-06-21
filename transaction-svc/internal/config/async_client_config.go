package config

import (
	"go-saga-pattern/commoner/utils"

	"github.com/hibiken/asynq"
)

func NewAsyncConfig() *asynq.Client {
	host := utils.GetEnv("REDIS_HOST")
	port := utils.GetEnv("REDIS_PORT")
	address := host + ":" + port

	return asynq.NewClient(asynq.RedisClientOpt{Addr: address})
}
