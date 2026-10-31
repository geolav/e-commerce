package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
)

// newPool создаёт пул соединений с PostgreSQL с ретраями.
//
// Почему ретраи:
//   - В Docker Compose контейнер с БД может стартовать дольше, чем сервис.
//     Без ретраев сервис упадёт при первом неудачном подключении.
//   - 10 попыток с интервалом 2 секунды дают 20 секунд на запуск БД — этого
//     достаточно для большинства окружений.
//
// Почему net.JoinHostPort:
//   - Безопасное соединение хоста и порта, учитывающее IPv6 адреса (где порт
//     нужно оборачивать в квадратные скобки). Простая конкатенация с ":"
//     работает только для IPv4.
func newPool(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	hostport := net.JoinHostPort(cfg.PG.Host, cfg.PG.Port)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&search_path=loms,public",
		cfg.PG.User, cfg.PG.Password, hostport, cfg.PG.DB,
	)

	const timeout = 3 * time.Second
	for i := 0; i < 10; i++ {
		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		pool, err := pgxpool.New(ctxTimeout, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctxTimeout); pingErr == nil {
				cancel()
				return pool, nil
			}
			pool.Close()
		}
		cancel()
		logger.Info("waiting for postgres...", zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}
	return nil, errors.New("cannot connect to postgres after retries")
}
