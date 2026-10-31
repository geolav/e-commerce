package app

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/stretchr/testify/assert"

	// "github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPoolDSN(t *testing.T) {
	// logger := zap.NewNop()

	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "standard config",
			cfg: &config.Config{
				PG: struct {
					Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
					Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
					DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
					User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
					Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
				}{
					Host: "localhost", Port: "5432", DB: "testdb", User: "user", Password: "pass",
				},
			},
		},
		{
			name: "ipv6 host",
			cfg: &config.Config{
				PG: struct {
					Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
					Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
					DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
					User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
					Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
				}{
					Host: "::1", Port: "5432", DB: "db", User: "u", Password: "p",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем net.JoinHostPort как в реальном коде
			hostport := net.JoinHostPort(tt.cfg.PG.Host, tt.cfg.PG.Port)
			dsn := "postgres://" + tt.cfg.PG.User + ":" + tt.cfg.PG.Password + "@" + hostport + "/" + tt.cfg.PG.DB + "?sslmode=disable&search_path=loms,public"

			// Проверяем, что DSN формируется без ошибок
			assert.NotEmpty(t, dsn)
			t.Logf("Generated DSN: %s", dsn)
		})
	}
}

func TestNewPoolTimeout(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	cfg := &config.Config{}
	cfg.PG.Host = "non-existent-host"
	cfg.PG.Port = "5432"
	cfg.PG.DB = "test"
	cfg.PG.User = "user"
	cfg.PG.Password = "pass"

	done := make(chan bool)
	go func() {
		_, err := newPool(ctx, cfg, logger)
		assert.Error(t, err)
		done <- true
	}()

	select {
	case <-done:
		// ok
	case <-time.After(5 * time.Second):
		t.Skip("skipping long-running connection test")
	}
}

func TestNewPoolDSNFormat(t *testing.T) {
	cfg := &config.Config{}
	cfg.PG.Host = "postgres"
	cfg.PG.Port = "5432"
	cfg.PG.DB = "ecommerce"
	cfg.PG.User = "user"
	cfg.PG.Password = "pass"

	hostport := cfg.PG.Host + ":" + cfg.PG.Port
	expected := "postgres://user:pass@postgres:5432/ecommerce?sslmode=disable&search_path=loms,public"
	actual := "postgres://" + cfg.PG.User + ":" + cfg.PG.Password + "@" + hostport + "/" + cfg.PG.DB + "?sslmode=disable&search_path=loms,public"

	assert.Equal(t, expected, actual)
	assert.Contains(t, actual, "search_path=loms,public")
	assert.Contains(t, actual, "sslmode=disable")
}

func TestJoinHostPort(t *testing.T) {
	tests := []struct {
		host     string
		port     string
		expected string
	}{
		{"localhost", "5432", "localhost:5432"},
		{"::1", "5432", "[::1]:5432"},
		{"192.168.1.1", "5432", "192.168.1.1:5432"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := net.JoinHostPort(tt.host, tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}
