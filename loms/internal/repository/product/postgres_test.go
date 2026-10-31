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
	tests := []struct {
		name    string
		pName   string
		price   uint32
		wantSKU uint32
	}{
		{"success", "Test", 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			mock.ExpectQuery(`INSERT INTO loms.products`).
				WithArgs(tt.pName, tt.price).
				WillReturnRows(pgxmock.NewRows([]string{"sku"}).AddRow(tt.wantSKU))

			sku, err := repo.CreateProduct(context.Background(), tt.pName, tt.price)
			require.NoError(t, err)
			require.Equal(t, tt.wantSKU, sku)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresGetProduct(t *testing.T) {
	tests := []struct {
		name      string
		sku       uint32
		mockSetup func(mock pgxmock.PgxPoolIface)
		wantName  string
		wantErr   error
	}{
		{
			name: "success",
			sku:  1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT name, price FROM loms.products`).
					WithArgs(uint32(1)).
					WillReturnRows(pgxmock.NewRows([]string{"name", "price"}).AddRow("Test", uint32(100)))
			},
			wantName: "Test",
		},
		{
			name: "not found",
			sku:  999,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT name, price FROM loms.products`).
					WithArgs(uint32(999)).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: entity.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			tt.mockSetup(mock)

			product, err := repo.GetProduct(context.Background(), tt.sku)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantName, product.Name)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresCreateProduct_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}
	mock.ExpectQuery(`INSERT INTO loms.products`).
		WithArgs("Test", uint32(100)).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.CreateProduct(context.Background(), "Test", 100)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetProduct_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}
	mock.ExpectQuery(`SELECT name, price FROM loms.products`).
		WithArgs(uint32(1)).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetProduct(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetStock_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}
	mock.ExpectQuery(`SELECT COALESCE`).
		WithArgs(uint32(1)).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetStock(context.Background(), 1)
	require.NoError(t, err) // GetStock возвращает 0 при ошибке
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresGetStock(t *testing.T) {
	tests := []struct {
		name      string
		sku       uint32
		mockSetup func(mock pgxmock.PgxPoolIface)
		wantCount uint64
	}{
		{
			name: "returns count",
			sku:  1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT COALESCE`).
					WithArgs(uint32(1)).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(uint64(10)))
			},
			wantCount: 10,
		},
		{
			name: "not found returns zero",
			sku:  999,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT COALESCE`).
					WithArgs(uint32(999)).
					WillReturnError(pgx.ErrNoRows)
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			tt.mockSetup(mock)

			count, err := repo.GetStock(context.Background(), tt.sku)
			require.NoError(t, err)
			require.Equal(t, tt.wantCount, count)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresStockOperations(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock pgxmock.PgxPoolIface)
		run       func(repo *postgresRepo, tx pgx.Tx) error
		wantErr   error
	}{
		{
			name: "SetStock success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO loms.stocks`).
					WithArgs(uint32(1), int64(50)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			run: func(repo *postgresRepo, _ pgx.Tx) error {
				return repo.SetStock(context.Background(), 1, 50)
			},
		},
		{
			name: "ReserveStock success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBeginTx(pgx.TxOptions{})
				mock.ExpectExec(`UPDATE loms.stocks SET count = count - `).
					WithArgs(int64(3), uint32(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			run: func(repo *postgresRepo, tx pgx.Tx) error {
				return repo.ReserveStock(context.Background(), tx, 1, 3)
			},
		},
		{
			name: "ReserveStock insufficient",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBeginTx(pgx.TxOptions{})
				mock.ExpectExec(`UPDATE loms.stocks SET count = count - `).
					WithArgs(int64(10), uint32(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				mock.ExpectCommit()
			},
			run: func(repo *postgresRepo, tx pgx.Tx) error {
				return repo.ReserveStock(context.Background(), tx, 1, 10)
			},
			wantErr: entity.ErrInsufficientStock,
		},
		{
			name: "ReleaseStock success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBeginTx(pgx.TxOptions{})
				mock.ExpectExec(`UPDATE loms.stocks SET count = count \+ `).
					WithArgs(int64(5), uint32(1)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			run: func(repo *postgresRepo, tx pgx.Tx) error {
				return repo.ReleaseStock(context.Background(), tx, 1, 5)
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

			var tx pgx.Tx
			// для операций с tx — создаём транзакцию через mock
			if tt.name != "SetStock success" {
				tx, err = mock.BeginTx(context.Background(), pgx.TxOptions{})
				require.NoError(t, err)
			}

			err = tt.run(repo, tx)

			if tx != nil {
				_ = tx.Commit(context.Background())
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
