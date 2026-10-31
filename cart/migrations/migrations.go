package migrations

import (
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// SetupPostgres накатывает миграции через goose, используя embed.FS.
//
// Почему embed, а не чтение с диска:
//
//  1. На лекции было рассказано, что миграции вкомпиливаются в бинарник на этапе сборки. Это исключает ошибки
//     "directory does not exist" в Docker-контейнере, где структура папок может
//     отличаться от локальной. И особенно полезно для CI.
//
// Почему goose:
//  1. Поддерживает версионирование миграций (up/down).
//  2. Работает как библиотека в коде.
//  3. Автоматически отслеживает, какие миграции уже применены.
func SetupPostgres(pool *pgxpool.Pool, logger *zap.Logger) {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatal("goose set dialect", zap.Error(err))
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Fatal("error running migrations", zap.Error(err))
	}
}
