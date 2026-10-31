package outbox

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestPostgresAdd(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "success",
			payload: map[string]interface{}{
				"order_id": int64(1),
				"user_id":  int64(2),
				"status":   "awaiting_payment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}

			mock.ExpectBeginTx(pgx.TxOptions{})
			mock.ExpectExec(`INSERT INTO loms.outbox`).
				WithArgs(pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			mock.ExpectCommit()

			tx, err := mock.BeginTx(context.Background(), pgx.TxOptions{})
			require.NoError(t, err)

			err = repo.Add(context.Background(), tx, tt.payload)
			require.NoError(t, err)

			require.NoError(t, tx.Commit(context.Background()))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresFetchBatch(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
		wantLen   int
	}{
		{
			name:      "returns events",
			batchSize: 10,
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}

			payload, _ := json.Marshal(map[string]interface{}{
				"order_id": int64(42),
				"user_id":  int64(7),
				"status":   "awaiting_payment",
			})

			mock.ExpectQuery(`UPDATE loms.outbox`).
				WithArgs(tt.batchSize, pgxmock.AnyArg()).
				WillReturnRows(pgxmock.NewRows([]string{"id", "payload", "attempts"}).
					AddRow(int64(1), payload, 1))

			events, err := repo.FetchBatch(context.Background(), tt.batchSize, time.Second)
			require.NoError(t, err)
			require.Len(t, events, tt.wantLen)
			require.Equal(t, float64(42), events[0].Payload["order_id"])
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresFetchBatch_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}

	mock.ExpectQuery(`UPDATE loms.outbox`).
		WithArgs(10, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "payload", "attempts"}))

	events, err := repo.FetchBatch(context.Background(), 10, time.Second)
	require.NoError(t, err)
	require.Empty(t, events)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresDelete_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &postgresRepo{pool: mock}
	mock.ExpectExec(`DELETE FROM loms.outbox`).
		WithArgs(int64(999)).
		WillReturnError(pgx.ErrNoRows)

	err = repo.Delete(context.Background(), 999)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresDelete(t *testing.T) {
	tests := []struct {
		name string
		id   int64
	}{
		{"success", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := &postgresRepo{pool: mock}
			mock.ExpectExec(`DELETE FROM loms.outbox`).
				WithArgs(tt.id).
				WillReturnResult(pgxmock.NewResult("DELETE", 1))

			err = repo.Delete(context.Background(), tt.id)
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
