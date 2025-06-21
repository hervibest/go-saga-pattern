package config

import (
	"fmt"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func NewJetStream(log logs.Log) nats.JetStreamContext {
	host := utils.GetEnv("NATS_HOST")
	port := utils.GetEnv("NATS_PORT")
	nc, err := nats.Connect(fmt.Sprintf("nats://%s:%s", host, port))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to nats server at %s:%s", host, port), zap.Error(err))
		return nil
	}

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal("Failed to create JetStream context", zap.Error(err))
		return nil
	}

	log.Info("Connected to JetStream", zap.String("host", host), zap.String("port", port))
	return js
}

func DeleteWebhookStream(js nats.JetStreamContext, log logs.Log) {
	err := js.DeleteStream("WEBHOOK_NOTIFY_STREAM")
	if err != nil {
		log.Error("failed to delete stream", zap.Error(err))
		return
	}
	log.Info("successfully deleted WEBHOOK_NOTIFY_STREAM")
}

func InitWebhookStream(js nats.JetStreamContext, log logs.Log) {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "WEBHOOK_NOTIFY_STREAM",
		Subjects: []string{"webhook.notify"},
		Storage:  nats.FileStorage,
	})

	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Fatal("failed to create stream", zap.Error(err))
	}
}
