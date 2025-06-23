package config

import (
	"go-saga-pattern/commoner/logs"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

func NewGocron(log logs.Log) gocron.Scheduler {
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local),
	)
	if err != nil {
		log.Fatal("Failed to create gocron scheduler", zap.Error(err))
	}

	return s
}
