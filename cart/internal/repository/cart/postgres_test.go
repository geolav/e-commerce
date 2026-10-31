package cart

import (
	"context"
	"fmt"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestPostgresCartRepository(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock pgxmock.PgxPoolIface)
		run       func(repo *postgresRepo) error
		wantErr   bool
	}{
		{
			name: "AddItem success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO cart.items`).
					WithArgs(int64(1), uint32(101), uint32(2)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			run: func(repo *postgresRepo) error {
				return repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 101, Count: 2})
			},
		},
		{
			name: "AddItem db error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO cart.items`).
					WithArgs(int64(1), uint32(101), uint32(2)).
					WillReturnError(fmt.Errorf("db error"))
			},
			run: func(repo *postgresRepo) error {
				return repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 101, Count: 2})
			},
			wantErr: true,
		},
		{
			name: "DeleteItem success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM cart.items WHERE user_id`).
					WithArgs(int64(1), uint32(101)).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			run: func(repo *postgresRepo) error {
				return repo.DeleteItem(context.Background(), 1, 101)
			},
		},
		{
			name: "ClearCart success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM cart.items WHERE user_id`).
					WithArgs(int64(1)).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			run: func(repo *postgresRepo) error {
				return repo.ClearCart(context.Background(), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			tt.mockSetup(mock)

			err = tt.run(repo)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresListItems(t *testing.T) {
	tests := []struct {
		name    string
		rows    [][]any
		wantLen int
	}{
		{
			name:    "returns items",
			rows:    [][]any{{uint32(101), uint32(2)}},
			wantLen: 1,
		},
		{
			name:    "empty cart",
			rows:    nil,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			rows := pgxmock.NewRows([]string{"sku", "count"})
			for _, r := range tt.rows {
				rows.AddRow(r...)
			}
			mock.ExpectQuery(`SELECT sku, count FROM cart.items`).
				WithArgs(int64(1)).
				WillReturnRows(rows)

			items, err := repo.ListItems(context.Background(), 1)
			require.NoError(t, err)
			require.Len(t, items, tt.wantLen)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
