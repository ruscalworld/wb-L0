package env

import "wb-l0/internal/config"

func ReadConfig() *config.Config {
	return &config.Config{
		Postgres: config.PostgresConnection{
			URL: requireEnv("POSTGRES_URL"),
		},
		Redis: config.RedisConnection{
			Address: envOrDefault("REDIS_ADDRESS", "127.0.0.1:6379"),
		},
		Server: config.Server{
			BindAddress: envOrDefault("BIND_ADDRESS", ":8080"),
		},
		Nats: config.NatsConnection{
			URL:     requireEnv("NATS_URL"),
			Subject: requireEnv("NATS_SUBJECT"),
		},
	}
}
