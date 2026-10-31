package cart

import (
	"context"
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

func TestPostgresAddItem(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`INSERT INTO cart.items`).
		WithArgs(int64(1), uint32(101), uint32(2)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 101, Count: 2})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresDeleteItem(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`DELETE FROM cart.items WHERE user_id`).
		WithArgs(int64(1), uint32(101)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteItem(context.Background(), 1, 101)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresListItems(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT sku, count FROM cart.items`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"sku", "count"}).
			AddRow(uint32(101), uint32(2)))

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint32(101), items[0].SKU)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresListItems_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT sku, count FROM cart.items`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"sku", "count"}))

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresClearCart(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`DELETE FROM cart.items WHERE user_id`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.ClearCart(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresAddItem_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`INSERT INTO cart.items`).
		WithArgs(int64(1), uint32(101), uint32(2)).
		WillReturnError(fmt.Errorf("db error"))

	err = repo.AddItem(context.Background(), 1, entity.OrderItem{
		SKU:   101,
		Count: 2,
	})

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
