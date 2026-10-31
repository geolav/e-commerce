package product

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

func TestPostgresCreateProduct(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`INSERT INTO loms.products`).
		WithArgs("Test", uint32(100)).
		WillReturnRows(pgxmock.NewRows([]string{"sku"}).AddRow(uint32(1)))

	sku, err := repo.CreateProduct(context.Background(), "Test", 100)
	require.NoError(t, err)
	require.Equal(t, uint32(1), sku)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetProduct_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT name, price FROM loms.products`).
		WithArgs(uint32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"name", "price"}).
			AddRow("Test", uint32(100)))

	product, err := repo.GetProduct(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "Test", product.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetProduct_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT name, price FROM loms.products`).
		WithArgs(uint32(999)).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetProduct(context.Background(), 999)
	require.ErrorIs(t, err, entity.ErrProductNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetStock(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT COALESCE`).
		WithArgs(uint32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(uint64(10)))

	count, err := repo.GetStock(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(10), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetStock_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`SELECT COALESCE`).
		WithArgs(uint32(999)).
		WillReturnError(pgx.ErrNoRows)

	count, err := repo.GetStock(context.Background(), 999)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresSetStock(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`INSERT INTO loms.stocks`).
		WithArgs(uint32(1), int64(50)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.SetStock(context.Background(), uint32(1), 50)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func newProductForTest(t *testing.T) (*entity.Product, error) {
	return &entity.Product{SKU: 1, Name: "Test", Price: 100}, nil
}

func TestPostgresReserveStock_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`UPDATE loms.stocks SET count = count - `).
		WithArgs(int64(3), uint32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.ReserveStock(context.Background(), 1, 3)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresReserveStock_InsufficientStock(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`UPDATE loms.stocks SET count = count - `).
		WithArgs(int64(10), uint32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = repo.ReserveStock(context.Background(), 1, 10)
	require.ErrorIs(t, err, entity.ErrInsufficientStock)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresReleaseStock_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`UPDATE loms.stocks SET count = count \+ `).
		WithArgs(int64(5), uint32(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.ReleaseStock(context.Background(), 1, 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
