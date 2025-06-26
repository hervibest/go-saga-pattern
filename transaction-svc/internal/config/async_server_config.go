package config

import "github.com/hibiken/asynq"

func NewAsynqServer() *asynq.Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: "127.0.0.1:6379",
			// Password:     "",               // jika pakai password
			// DB:           0,                // optional: default DB
			// TLSConfig:    &tls.Config{},    // jika pakai Redis TLS (bukan HTTPS ya!)
			// DialTimeout:  10 * time.Second, // increased timeout
			// ReadTimeout:  10 * time.Second,
			// WriteTimeout: 10 * time.Second,
		},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 2,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)
	return srv
}
