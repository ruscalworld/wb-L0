package config

type Config struct {
	Postgres PostgresConnection
	Redis    RedisConnection
	Nats     NatsConnection
	Server   Server
}

type PostgresConnection struct {
	URL string
}

type RedisConnection struct {
	Address string
}

type Server struct {
	BindAddress string
}

type NatsConnection struct {
	URL     string
	Subject string
}
