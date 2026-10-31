package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
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

func (r *postgresRepo) ReserveStock(ctx context.Context, tx pgx.Tx, sku uint32, count uint64) error {
	result, err := tx.Exec(ctx,
		`UPDATE loms.stocks SET count = count - $1 WHERE sku = $2 AND count >= $1`, int64(count), sku)
	if err != nil {
		return fmt.Errorf("error reserve stock: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrInsufficientStock
	}
	return nil
}

func (r *postgresRepo) ReleaseStock(ctx context.Context, tx pgx.Tx, sku uint32, count uint64) error {
	_, err := tx.Exec(ctx,
		`UPDATE loms.stocks SET count = count + $1 WHERE sku = $2`, int64(count), sku)
	if err != nil {
		return fmt.Errorf("error release stock: %w", err)
	}
	return nil
}

func (r *postgresRepo) CreateProduct(ctx context.Context, name string, price uint32) (uint32, error) {
	var sku uint32
	err := r.pool.QueryRow(ctx,
		`INSERT INTO loms.products (name, price) VALUES ($1, $2) RETURNING sku`,
		name, price,
	).Scan(&sku)
	if err != nil {
		return 0, fmt.Errorf("error create product: %w", err)
	}
	return sku, nil
}

func (r *postgresRepo) GetProduct(ctx context.Context, sku uint32) (*entity.Product, error) {
	product := &entity.Product{SKU: sku}
	err := r.pool.QueryRow(ctx,
		`SELECT name, price FROM loms.products WHERE sku = $1`,
		sku,
	).Scan(&product.Name, &product.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProductNotFound
		}
		return nil, fmt.Errorf("error get product: %w", err)
	}
	return product, nil
}

func (r *postgresRepo) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	var count uint64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(count, 0) FROM loms.stocks WHERE sku = $1`,
		sku,
	).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error get stock: %w", err)
	}
	return count, nil
}

func (r *postgresRepo) SetStock(ctx context.Context, sku uint32, count uint64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO loms.stocks (sku, count) VALUES ($1, $2)
         ON CONFLICT (sku) DO UPDATE SET count = $2`,
		sku, int64(count),
	)
	if err != nil {
		return fmt.Errorf("error set stock: %w", err)
	}
	return nil
}
