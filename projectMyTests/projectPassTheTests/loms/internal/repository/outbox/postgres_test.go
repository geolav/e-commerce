package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestPostgresAdd(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`INSERT INTO loms.outbox`).
		WithArgs(int64(1), int64(2), "awaiting_payment").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Add(context.Background(), 1, 2, "awaiting_payment")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresFetchBatch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`UPDATE loms.outbox`).
		WithArgs(10, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "order_id", "user_id", "status", "attempts"}).
			AddRow(int64(1), int64(42), int64(7), "awaiting_payment", 1))

	events, err := repo.FetchBatch(context.Background(), 10, time.Second)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, int64(42), events[0].OrderID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectExec(`DELETE FROM loms.outbox`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
