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
	tests := []struct {
		name   string
		order  *entity.Order
		wantID int64
	}{
		{
			name:   "success",
			order:  &entity.Order{UserID: 1, Items: []entity.OrderItem{{SKU: 101, Count: 2}}},
			wantID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			mock.ExpectCommit()

			tx, err := mock.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			require.NoError(t, err)

			id, err := repo.CreateOrder(context.Background(), tx, tt.order)
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)

			require.NoError(t, tx.Commit(context.Background()))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresGetOrder(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		orderID    int64
		mockSetup  func(mock pgxmock.PgxPoolIface)
		wantUserID int64
		wantErr    error
	}{
		{
			name:    "success",
			orderID: 1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT user_id, status, created_at, updated_at FROM loms.orders`).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "status", "created_at", "updated_at"}).
						AddRow(int64(1), entity.OrderStatusAwaitingPay, now, now))
				mock.ExpectQuery(`SELECT sku, count FROM loms.order_items`).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"sku", "count"}).
						AddRow(uint32(101), uint32(2)))
			},
			wantUserID: 1,
		},
		{
			name:    "not found",
			orderID: 999,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT user_id, status, created_at, updated_at FROM loms.orders`).
					WithArgs(int64(999)).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: entity.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			tt.mockSetup(mock)

			order, err := repo.GetOrder(context.Background(), tt.orderID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUserID, order.UserID)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresSetOrderStatus(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int64
		newStatus entity.OrderStatus
		mockSetup func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name:      "success awaiting->paid",
			orderID:   1,
			newStatus: entity.OrderStatusPayed,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
				mock.ExpectQuery(`SELECT status, user_id FROM loms.orders WHERE id`).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"status", "user_id"}).
						AddRow(entity.OrderStatusAwaitingPay, int64(1)))
				mock.ExpectExec(`UPDATE loms.orders SET status`).
					WithArgs("paid", int64(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
		},
		{
			name:      "not found",
			orderID:   999,
			newStatus: entity.OrderStatusPayed,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
				mock.ExpectQuery(`SELECT status, user_id FROM loms.orders WHERE id`).
					WithArgs(int64(999)).
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: entity.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			tt.mockSetup(mock)

			tx, err := mock.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
			require.NoError(t, err)

			err = repo.SetOrderStatus(context.Background(), tx, tt.orderID, tt.newStatus)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				_ = tx.Rollback(context.Background())
			} else {
				require.NoError(t, err)
				_ = tx.Commit(context.Background())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestStatusToString(t *testing.T) {
	tests := []struct {
		input    entity.OrderStatus
		expected string
	}{
		{entity.OrderStatusNew, "new"},
		{entity.OrderStatusFailed, "failed"},
		{entity.OrderStatusCancelled, "cancelled"},
		{entity.OrderStatusPayed, "paid"},
		{entity.OrderStatusAwaitingPay, "awaiting_payment"},
		{entity.OrderStatus("unknown"), "new"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, statusToString(tt.input))
		})
	}
}
