package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const New = "new"

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

func (r *postgresRepo) CreateOrder(ctx context.Context, tx pgx.Tx, order *entity.Order) (int64, error) {
	// убери создание транзакции — она приходит снаружи
	var orderID int64
	err := tx.QueryRow(ctx,
		`INSERT INTO loms.orders (user_id, status) VALUES ($1, $2) RETURNING id`,
		order.UserID, "awaiting_payment").Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("error insert order: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO loms.order_items (order_id, sku, count) VALUES ($1, $2, $3)`,
			orderID, item.SKU, item.Count)
		if err != nil {
			return 0, fmt.Errorf("error insert order item: %w", err)
		}
	}

	order.ID = orderID
	return orderID, nil
}

func (r *postgresRepo) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	order := &entity.Order{ID: orderID}
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, status, created_at, updated_at FROM loms.orders WHERE id = $1`, orderID,
	).Scan(&order.UserID, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, fmt.Errorf("error get order: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT sku, count FROM loms.order_items WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("error get order_items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.SKU, &item.Count); err != nil {
			return nil, fmt.Errorf("error scan order_item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	return order, nil
}

// SetOrderStatus атомарно меняет статус заказа с проверкой текущего состояния.
//
// Используется Repeatable Read вместо Serializable: Serializable даёт больше ложных конфликтов и требует ретраев, замедляя работу.
// Repeatable Read + FOR UPDATE достаточно для блокировки строки до конца транзакции.
//
// WHERE status = $3 + RowsAffected: дополнительная защита от гонок. Даже если две транзакции прочитали одинаковый currentStatus,
// UPDATE затронет только одну строку, вторая получит 0 rows affected.
//func (r *postgresRepo) SetOrderStatus(ctx context.Context, orderID int64, status entity.OrderStatus) error {
//	for i := 0; i < 5; i++ {
//		err := r.setOrderStatusOnce(ctx, orderID, status)
//		if err == nil {
//			return nil
//		}
//		var pgErr *pgconn.PgError
//		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
//			continue
//		}
//		return err
//	}
//	return errors.New("error set order status after retries")
//}
//
//func (r *postgresRepo) setOrderStatusOnce(ctx context.Context, orderID int64, status entity.OrderStatus) error {
//	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
//	if err != nil {
//		return fmt.Errorf("error begin tx: %w", err)
//	}
//	defer func() { _ = tx.Rollback(ctx) }()
//
//	var currentStatus entity.OrderStatus
//	err = tx.QueryRow(ctx,
//		`SELECT status FROM loms.orders WHERE id = $1 FOR UPDATE`, orderID,
//	).Scan(&currentStatus)
//	if err != nil {
//		if errors.Is(err, pgx.ErrNoRows) {
//			return entity.ErrOrderNotFound
//		}
//		return fmt.Errorf("error get order_status: %w", err)
//	}
//
//	if status == entity.OrderStatusPayed && currentStatus != entity.OrderStatusAwaitingPay {
//		return entity.ErrInvalidStatus
//	}
//	if status == entity.OrderStatusCancelled &&
//		(currentStatus == entity.OrderStatusPayed || currentStatus == entity.OrderStatusCancelled) {
//		return entity.ErrInvalidStatus
//	}
//
//	result, err := tx.Exec(ctx,
//		`UPDATE loms.orders SET status = $1, updated_at = NOW()
//		 WHERE id = $2 AND status = $3`,
//		statusToString(status), orderID, statusToString(currentStatus))
//	if err != nil {
//		return fmt.Errorf("error update order_status: %w", err)
//	}
//	if result.RowsAffected() == 0 {
//		return entity.ErrInvalidStatus
//	}
//	_, err = tx.Exec(ctx,
//		`INSERT INTO loms.outbox (order_id, user_id, status)
//     SELECT $1, user_id, $2 FROM loms.orders WHERE id = $1`,
//		orderID, statusToString(status),
//	)
//	if err != nil {
//		return fmt.Errorf("error insert outbox: %w", err)
//	}
//
//	return tx.Commit(ctx)
//}

func (r *postgresRepo) SetOrderStatus(ctx context.Context, tx pgx.Tx, orderID int64, status entity.OrderStatus) error {
	var currentStatus entity.OrderStatus
	var userID int64
	err := tx.QueryRow(ctx,
		`SELECT status, user_id FROM loms.orders WHERE id = $1 FOR UPDATE`, orderID,
	).Scan(&currentStatus, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ErrOrderNotFound
		}
		return fmt.Errorf("error get order_status: %w", err)
	}

	if status == entity.OrderStatusPayed && currentStatus != entity.OrderStatusAwaitingPay {
		return entity.ErrInvalidStatus
	}
	if status == entity.OrderStatusCancelled &&
		(currentStatus == entity.OrderStatusPayed || currentStatus == entity.OrderStatusCancelled) {
		return entity.ErrInvalidStatus
	}

	_, err = tx.Exec(ctx,
		`UPDATE loms.orders SET status = $1 WHERE id = $2`,
		statusToString(status), orderID)
	if err != nil {
		return fmt.Errorf("error update order_status: %w", err)
	}

	//payload := fmt.Sprintf(`{"order_id":%d,"user_id":%d,"status":"%s"}`, orderID, userID, statusToString(status))
	//_, err = tx.Exec(ctx,
	//	`INSERT INTO loms.outbox (payload) VALUES ($1::jsonb)`, payload)
	//if err != nil {
	//	return fmt.Errorf("error insert outbox: %w", err)
	//}

	return nil
}

func statusToString(status entity.OrderStatus) string {
	switch status {
	case entity.OrderStatusCancelled:
		return "cancelled"
	case entity.OrderStatusPayed:
		return "paid"
	case entity.OrderStatusAwaitingPay:
		return "awaiting_payment"
	case entity.OrderStatusFailed:
		return "failed"
	case entity.OrderStatusNew:
		return New
	default:
		return New
	}
}
