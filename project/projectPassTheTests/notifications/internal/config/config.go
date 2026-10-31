package config

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50053"`
	}

	Clients struct {
		CallbackAddr string `env:"CALLBACK_ADDR" envDefault:""`
	}

	Kafka struct {
		Brokers       string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
		Topic         string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
		ConsumerGroup string `env:"KAFKA_CONSUMER_GROUP" envDefault:"notifications"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	//if cfg.Clients.CallbackAddr == "" {
	//	// Проверяем, запущены ли мы в Docker
	//	if _, err := os.Stat("/.dockerenv"); err == nil {
	//		cfg.Clients.CallbackAddr = "integration-tests:8080/callback"
	//	} else {
	//		cfg.Clients.CallbackAddr = "localhost:8080/callback"
	//	}
	//}

	return &cfg, err
}
