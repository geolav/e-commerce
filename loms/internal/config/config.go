package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	GRPC struct {
		Port        string `env:"GRPC_PORT" envDefault:"50052"`
		GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
	}

	PG struct {
		Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
		Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
		DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
		User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
		Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
	}

	// Clients struct {
	//	NotificationsGrpcAddr string `env:"NOTIFICATIONS_GRPC_ADDR" envDefault:"localhost:50053"`
	//}

	Outbox struct {
		Workers       int           `env:"OUTBOX_WORKERS" envDefault:"1"`
		BatchSize     int           `env:"OUTBOX_BATCH_SIZE" envDefault:"10"`
		FetchPeriod   time.Duration `env:"OUTBOX_FETCH_PERIOD" envDefault:"20ms"`
		InProgressTTL time.Duration `env:"OUTBOX_IN_PROGRESS_TTL" envDefault:"40s"`
	}

	Kafka struct {
		Brokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
		Topic   string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
