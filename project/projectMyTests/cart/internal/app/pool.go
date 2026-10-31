package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/config"
)

func newPool(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	hostport := net.JoinHostPort(cfg.PG.Host, cfg.PG.Port)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s",
		cfg.PG.User, cfg.PG.Password, hostport, cfg.PG.DB,
	)

	for i := 0; i < 10; i++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			}
		}
		logger.Info("waiting for postgres...", zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	return nil, errors.New("failed to connect to postgres")
}
