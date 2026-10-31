package order

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

func TestPostgresCreateOrder(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	mock.ExpectQuery(`INSERT INTO loms.orders`).
		WithArgs(int64(1), "awaiting_payment").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec(`INSERT INTO loms.order_items`).
		WithArgs(int64(1), uint32(101), uint32(2)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(`INSERT INTO loms.outbox`).
		WithArgs(int64(1), int64(1), "awaiting_payment").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	order := &entity.Order{
		UserID: 1,
		Items:  []entity.OrderItem{{SKU: 101, Count: 2}},
	}

	id, err := repo.CreateOrder(context.Background(), order)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetOrder_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT user_id, status, created_at, updated_at FROM loms.orders`).
		WithArgs(int64(999)).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetOrder(context.Background(), 999)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetOrder_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	now := time.Now()
	mock.ExpectQuery(`SELECT user_id, status, created_at, updated_at FROM loms.orders`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "status", "created_at", "updated_at"}).
			AddRow(int64(1), entity.OrderStatusAwaitingPay, now, now))

	mock.ExpectQuery(`SELECT sku, count FROM loms.order_items`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"sku", "count"}).
			AddRow(uint32(101), uint32(2)))

	order, err := repo.GetOrder(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), order.UserID)
	require.Len(t, order.Items, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSetOrderStatus_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	mock.ExpectQuery(`SELECT status FROM loms.orders WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"status"}).
			AddRow(entity.OrderStatusAwaitingPay))
	mock.ExpectExec(`UPDATE loms.orders SET status`).
		WithArgs(string(entity.OrderStatusPayed), int64(1), string(entity.OrderStatusAwaitingPay)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	err = repo.SetOrderStatus(context.Background(), 1, entity.OrderStatusPayed)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSetOrderStatus_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	mock.ExpectQuery(`SELECT status FROM loms.orders WHERE id`).
		WithArgs(int64(999)).
		WillReturnError(pgx.ErrNoRows)
	mock.ExpectRollback()

	err = repo.SetOrderStatus(context.Background(), 999, entity.OrderStatusPayed)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStatusToString(t *testing.T) {
	require.Equal(t, "new", statusToString(entity.OrderStatusNew))
	require.Equal(t, "failed", statusToString(entity.OrderStatusFailed))
	require.Equal(t, "cancelled", statusToString(entity.OrderStatusCancelled))
	require.Equal(t, "paid", statusToString(entity.OrderStatusPayed))
	require.Equal(t, "awaiting_payment", statusToString(entity.OrderStatusAwaitingPay))
	require.Equal(t, "new", statusToString(entity.OrderStatus("unknown")))
}
