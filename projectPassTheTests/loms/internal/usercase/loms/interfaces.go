package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/jackc/pgx/v5"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, tx pgx.Tx, order *entity.Order) (int64, error)
	GetOrder(ctx context.Context, orderID int64) (*entity.Order, error)
	SetOrderStatus(ctx context.Context, tx pgx.Tx, orderID int64, status entity.OrderStatus) error
}

type ProductRepository interface {
	ReserveStock(ctx context.Context, tx pgx.Tx, sku uint32, count uint64) error
	ReleaseStock(ctx context.Context, tx pgx.Tx, sku uint32, count uint64) error
	GetStock(ctx context.Context, sku uint32) (uint64, error)            // чтение без транзакции
	GetProduct(ctx context.Context, sku uint32) (*entity.Product, error) // чтение
}

type OutboxRepository interface {
	Add(ctx context.Context, tx pgx.Tx, payload map[string]interface{}) error
	// FetchBatch и Delete остаются в outbox/postgres.go без tx
}
