//package outbox
//
//import (
//	"context"
//	"fmt"
//	"time"
//
//	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
//	"github.com/jackc/pgx/v5"
//	"github.com/jackc/pgx/v5/pgconn"
//)
//
//type pgxPool interface {
//	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
//	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
//}
//
//type postgresRepo struct {
//	pool pgxPool
//}
//
//func NewPostgresRepo(pool pgxPool) *postgresRepo {
//	return &postgresRepo{pool: pool}
//}
//func (r *postgresRepo) Add(ctx context.Context, orderID, userID int64, status string) error {
//	_, err := r.pool.Exec(ctx,
//		`INSERT INTO loms.outbox (order_id, user_id, status) VALUES ($1, $2, $3)`,
//		orderID, userID, status,
//	)
//	return err
//}
//
//func (r *postgresRepo) FetchBatch(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]entity.OutboxEvent, error) {
//	rows, err := r.pool.Query(ctx, `
//        UPDATE loms.outbox
//        SET in_progress_at = NOW(), attempts = attempts + 1
//        WHERE id IN (
//            SELECT id FROM loms.outbox
//            WHERE in_progress_at IS NULL
//               OR in_progress_at < NOW() - $2::interval
//            ORDER BY id
//            LIMIT $1
//            FOR UPDATE SKIP LOCKED
//        )
//        RETURNING id, order_id, user_id, status, attempts
//    `, batchSize, fmt.Sprintf("%f seconds", inProgressTTL.Seconds()))
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var events []entity.OutboxEvent
//	for rows.Next() {
//		var e entity.OutboxEvent
//		if err := rows.Scan(&e.ID, &e.OrderID, &e.UserID, &e.Status, &e.Attempts); err != nil {
//			return nil, err
//		}
//		events = append(events, e)
//	}
//	return events, nil
//}
//
//func (r *postgresRepo) Delete(ctx context.Context, id int64) error {
//	_, err := r.pool.Exec(ctx, `DELETE FROM loms.outbox WHERE id = $1`, id)
//	return err
//}

package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxPool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type postgresRepo struct {
	pool pgxPool
}

func NewPostgresRepo(pool pgxPool) *postgresRepo {
	return &postgresRepo{pool: pool}
}

// Add принимает tx (будет передан из usecase)
func (r *postgresRepo) Add(ctx context.Context, tx pgx.Tx, payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	fmt.Printf("ADDING OUTBOX EVENT: %s\n", b)

	_, err = tx.Exec(ctx, `INSERT INTO loms.outbox (payload) VALUES ($1)`, b)
	return err
}

//// FetchBatch использует пул (не транзакцию)
//func (r *postgresRepo) FetchBatch(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]entity.OutboxEvent, error) {
//	rows, err := r.pool.Query(ctx, `
//        UPDATE loms.outbox
//        SET in_progress_at = NOW(), attempts = attempts + 1
//        WHERE id IN (
//            SELECT id FROM loms.outbox
//            WHERE in_progress_at IS NULL
//               OR in_progress_at < NOW() - $2::interval
//            ORDER BY id
//            LIMIT $1
//            FOR UPDATE SKIP LOCKED
//        )
//        RETURNING id, payload, attempts
//    `, batchSize, fmt.Sprintf("%f seconds", inProgressTTL.Seconds()))
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var events []entity.OutboxEvent
//	for rows.Next() {
//		var e entity.OutboxEvent
//		var payloadBytes []byte
//		if err := rows.Scan(&e.ID, &payloadBytes, &e.Attempts); err != nil {
//			return nil, err
//		}
//		if err := json.Unmarshal(payloadBytes, &e.Payload); err != nil {
//			return nil, fmt.Errorf("unmarshal payload: %w", err)
//		}
//		events = append(events, e)
//	}
//	return events, nil
//}

func (r *postgresRepo) FetchBatch(
	ctx context.Context,
	batchSize int,
	inProgressTTL time.Duration,
) ([]entity.OutboxEvent, error) {

	rows, err := r.pool.Query(ctx, `
    UPDATE loms.outbox
    SET
        in_progress_at = NOW(),
        attempts = attempts + 1
    WHERE id IN (
        SELECT id
        FROM loms.outbox
        WHERE (in_progress_at IS NULL OR in_progress_at < NOW() - $2::interval)
          AND attempts < 100
        ORDER BY id
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    )
    RETURNING id, payload, attempts
`, batchSize, fmt.Sprintf("%f seconds", inProgressTTL.Seconds()))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []entity.OutboxEvent

	for rows.Next() {
		var e entity.OutboxEvent
		var payloadBytes []byte

		if err := rows.Scan(
			&e.ID,
			&payloadBytes,
			&e.Attempts,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadBytes, &e.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}

		events = append(events, e)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return events, nil
}

func (r *postgresRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM loms.outbox WHERE id = $1`, id)
	return err
}
