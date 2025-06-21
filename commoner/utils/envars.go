package utils

import (
	"go-saga-pattern/commoner/logs"
	"os"
	"sync"

	vault "github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	vaultConfig struct {
		Host   string
		Port   string
		Auth   string
		Token  string
		Engine string
		Path   string
	}
	vaultClient *vault.Client
	vaultOnce   sync.Once
)

type StringArrayMarshaler []string

func (s StringArrayMarshaler) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, v := range s {
		enc.AppendString(v)
	}
	return nil
}

func logFailure(key string, sources []string) {
	logger, _ := logs.NewLogger()
	logger.Error("Failed to get key-value pair", zap.Array("failed sources", StringArrayMarshaler(sources)), zap.String("key", key))
}

func GetEnv(key string) string {
	var failedSources []string

	if value := getOSEnv(key, &failedSources); value != "" {
		return value
	}

	if value := getDotEnv(key, &failedSources); value != "" {
		return value
	}

	logFailure(key, failedSources)
	return ""
}

func getOSEnv(key string, failedSources *[]string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	*failedSources = append(*failedSources, "os")
	return ""
}

func getDotEnv(key string, failedSources *[]string) string {
	if err := godotenv.Load("../../.env"); err != nil {
		*failedSources = append(*failedSources, ".env")
		return ""
	}
	return getOSEnv(key, failedSources)
}
