package cart

import (
	"context"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxPool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type postgresRepo struct {
	pool pgxPool
}

func NewPostgresRepo(pool pgxPool) *postgresRepo {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) AddItem(ctx context.Context, userID int64, item entity.OrderItem) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO cart.items (user_id, sku, count) VALUES ($1, $2, $3) ON CONFLICT (user_id, sku) DO UPDATE SET count = cart.items.count + $3`,
		userID, item.SKU, item.Count)
	if err != nil {
		return fmt.Errorf("error add item: %w", err)
	}
	return nil
}

func (r *postgresRepo) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM cart.items WHERE user_id = $1 AND sku = $2`, userID, sku)
	if err != nil {
		return fmt.Errorf("error delete item: %w", err)
	}
	return nil
}

func (r *postgresRepo) ListItems(ctx context.Context, userID int64) ([]entity.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT sku, count FROM cart.items WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("error list items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.SKU, &item.Count); err != nil {
			return nil, fmt.Errorf("error scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepo) ClearCart(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM cart.items WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("error clear cart: %w", err)
	}
	return nil
}
