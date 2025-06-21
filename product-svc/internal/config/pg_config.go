package config

import (
	"context"
	"fmt"
	"go-saga-pattern/commoner/logs"
	"go-saga-pattern/commoner/utils"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

var (
	host     = utils.GetEnv("DB_HOST")
	port     = utils.GetEnv("DB_PORT")
	username = utils.GetEnv("DB_USERNAME")
	password = utils.GetEnv("DB_PASSWORD")
	dbName   = utils.GetEnv("DB_NAME")
	minConns = utils.GetEnv("DB_MIN_CONNS")
	maxConns = utils.GetEnv("DB_MAX_CONNS")
	sslMode  = utils.GetEnv("DB_SSL_MODE")
	timeZone = utils.GetEnv("DB_TIMEZONE")
)

func NewPostgresDatabase() *pgxpool.Pool {
	logger, _ := logs.NewLogger()
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s&timezone=%s", username, password, host, port, dbName, sslMode, timeZone)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Panic("Failed to parse configuration dsn " + dsn)
	}

	minConnsInt, err := strconv.Atoi(minConns)
	if err != nil {
		logger.Panic("DB_MIN_CONNS expected to be integer minimum connections " + minConns)
	}
	maxConnsInt, err := strconv.Atoi(maxConns)
	if err != nil {
		logger.Panic("DB_MAX_CONNS expected to be integer maximum connections" + maxConns)
	}

	poolConfig.MinConns = int32(minConnsInt)
	poolConfig.MaxConns = int32(maxConnsInt)
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Panic("Failed to apply pool configuration dsn " + dsn)
	}

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pool.Ping(c); err != nil {
		logger.Panic("Failed to connect to db  ", zap.Error(err))
	}

	return pool
}
