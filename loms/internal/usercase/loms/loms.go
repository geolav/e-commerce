package loms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/loms_mock.go -package=mocks

const (
	fieldOrderID = "order_id"
	fieldUserID  = "user_id"
	fieldStatus  = "status"
	statusAwait  = "awaiting_payment"
	timeDuration = 50
)

type lomsService struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
	outboxRepo  OutboxRepository
	db          TxBeginner
}

func NewLomsService(orderRepo OrderRepository, productRepo ProductRepository, outboxRepo OutboxRepository, db TxBeginner) *lomsService {
	return &lomsService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		db:          db,
	}
}

func (s *lomsService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		orderID, err := s.createOrderOnce(ctx, userID, items)
		if err == nil {
			return orderID, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			time.Sleep(time.Duration(timeDuration*(i+1)) * time.Millisecond)
			continue
		}
		return 0, err
	}
	return 0, fmt.Errorf("failed to create order after %d retries", maxRetries)
}

func (s *lomsService) createOrderOnce(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, item := range items {
		if err = s.productRepo.ReserveStock(ctx, tx, item.SKU, uint64(item.Count)); err != nil {
			return 0, err
		}
	}

	order := &entity.Order{UserID: userID, Items: items}
	orderID, err := s.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		fieldOrderID: orderID,
		fieldUserID:  userID,
		fieldStatus:  statusAwait,
	}
	if err = s.outboxRepo.Add(ctx, tx, payload); err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return orderID, nil
}

func (s *lomsService) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}

func (s *lomsService) PayOrder(ctx context.Context, orderID int64) error {
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		err := s.payOrderOnce(ctx, orderID)
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			time.Sleep(time.Duration(timeDuration*(i+1)) * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("failed to pay order after %d retries", maxRetries)
}

func (s *lomsService) payOrderOnce(ctx context.Context, orderID int64) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err = s.orderRepo.SetOrderStatus(ctx, tx, orderID, entity.OrderStatusPayed); err != nil {
		return err
	}

	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		fieldOrderID: orderID,
		fieldUserID:  order.UserID,
		fieldStatus:  "paid",
	}
	if err = s.outboxRepo.Add(ctx, tx, payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *lomsService) CancelOrder(ctx context.Context, orderID int64) error {
	const maxRetries = 7
	for i := 0; i < maxRetries; i++ {
		err := s.cancelOrderOnce(ctx, orderID)
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "40001" {
			time.Sleep(time.Duration(timeDuration*(i+1)) * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("failed to cancel order after %d retries", maxRetries)
}

func (s *lomsService) cancelOrderOnce(ctx context.Context, orderID int64) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if err = s.orderRepo.SetOrderStatus(ctx, tx, orderID, entity.OrderStatusCancelled); err != nil {
		return err
	}

	for _, item := range order.Items {
		if err = s.productRepo.ReleaseStock(ctx, tx, item.SKU, uint64(item.Count)); err != nil {
			return err
		}
	}

	payload := map[string]interface{}{
		fieldOrderID: orderID,
		fieldUserID:  order.UserID,
		fieldStatus:  "cancelled",
	}
	if err = s.outboxRepo.Add(ctx, tx, payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
