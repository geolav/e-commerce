//package loms
//
////go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks
//
//import (
//	"context"
//	"errors"
//
//	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
//)
//
//// Пакет loms содержит бизнес-логику управления заказами.
////
//// CreateOrder использует ReserveStock вместо GetStock+SetStock.
//// Это атомарная операция (UPDATE ... WHERE count >= $1), которая гарантирует, что при конкурентных заказах сток не уйдёт в минус.
//// Подход с GetStock требовал бы дополнительной синхронизации между чтением и записью.
////
//// CancelOrder выполняет SetOrderStatus до ReleaseStock.
//// Это гарантирует, что при конкурентных отменах сток вернётся ровно один раз — только тот запрос, который выиграл "гонку" за смену статуса, освободит сток.
////
//// PayOrder не требует отдельной проверки статуса, потому что SetOrderStatus
//// уже проверяет текущий статус через WHERE status = $3 и возвращает ErrInvalidStatus, если заказ не в awaiting_payment.
////
//// Ошибки мапятся в gRPC-коды на уровне контроллера (NotFound, FailedPrecondition, Internal), а не здесь. Это разделяет бизнес-логику и транспортный слой.
//
//var (
//	ErrOrderNotFound     = entity.ErrOrderNotFound
//	ErrInsufficientStock = entity.ErrInsufficientStock
//	ErrInvalidStatus     = entity.ErrInvalidStatus
//)
//
//type (
//	orderRepository interface {
//		CreateOrder(ctx context.Context, order *entity.Order) (int64, error)
//		GetOrder(ctx context.Context, orderID int64) (*entity.Order, error)
//		SetOrderStatus(ctx context.Context, orderID int64, status entity.OrderStatus) error
//	}
//
//	stocksRepository interface {
//		GetStock(ctx context.Context, sku uint32) (uint64, error)
//		SetStock(ctx context.Context, sku uint32, count uint64) error
//		ReserveStock(ctx context.Context, sku uint32, count uint64) error
//		ReleaseStock(ctx context.Context, sku uint32, count uint64) error
//	}
//)
//
//func NewLomsService(
//	orderRepository orderRepository,
//	stocksRepository stocksRepository,
//) *lomsService {
//	return &lomsService{
//		orderRepository:  orderRepository,
//		stocksRepository: stocksRepository,
//	}
//}
//
//type lomsService struct {
//	orderRepository  orderRepository
//	stocksRepository stocksRepository
//}
//
//func (s *lomsService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
//	for _, item := range items {
//		if err := s.stocksRepository.ReserveStock(ctx, item.SKU, uint64(item.Count)); err != nil {
//			if errors.Is(err, ErrInsufficientStock) {
//				return 0, ErrInsufficientStock
//			}
//			return 0, err
//		}
//	}
//	order := &entity.Order{
//		Items:  items,
//		UserID: userID,
//		Status: entity.OrderStatusAwaitingPay,
//	}
//	return s.orderRepository.CreateOrder(ctx, order)
//}
//
//func (s *lomsService) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
//	order, err := s.orderRepository.GetOrder(ctx, orderID)
//	if err != nil {
//		return nil, ErrOrderNotFound
//	}
//	return order, nil
//}
//
//func (s *lomsService) PayOrder(ctx context.Context, orderID int64) error {
//	err := s.orderRepository.SetOrderStatus(ctx, orderID, entity.OrderStatusPayed)
//	if err != nil {
//		if errors.Is(err, entity.ErrOrderNotFound) {
//			return ErrOrderNotFound
//		}
//		if errors.Is(err, entity.ErrInvalidStatus) {
//			return ErrInvalidStatus
//		}
//		return err
//	}
//	return nil
//}
//
//func (s *lomsService) CancelOrder(ctx context.Context, orderID int64) error {
//	// сначала атомарно меняем статус — только один cancel пройдёт
//	err := s.orderRepository.SetOrderStatus(ctx, orderID, entity.OrderStatusCancelled)
//	if err != nil {
//		if errors.Is(err, entity.ErrOrderNotFound) {
//			return ErrOrderNotFound
//		}
//		if errors.Is(err, entity.ErrInvalidStatus) {
//			return ErrInvalidStatus
//		}
//		return err
//	}
//
//	// статус успешно сменён — теперь отпускаем сток
//	order, err := s.orderRepository.GetOrder(ctx, orderID)
//	if err != nil {
//		// статус уже Cancelled, но сток не освобождён — логируем и выходим
//		return err
//	}
//	for _, item := range order.Items {
//		if err := s.stocksRepository.ReleaseStock(ctx, item.SKU, uint64(item.Count)); err != nil {
//			return err
//		}
//	}
//	return nil
//}

package loms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type lomsService struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
	outboxRepo  OutboxRepository
	db          *pgxpool.Pool
}

func NewLomsService(orderRepo OrderRepository, productRepo ProductRepository, outboxRepo OutboxRepository, db *pgxpool.Pool) *lomsService {
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
			time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
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
	defer tx.Rollback(ctx)

	for _, item := range items {
		if err := s.productRepo.ReserveStock(ctx, tx, item.SKU, uint64(item.Count)); err != nil {
			return 0, err
		}
	}

	order := &entity.Order{UserID: userID, Items: items}
	orderID, err := s.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		return 0, err
	}

	payload := map[string]interface{}{
		"order_id": orderID,
		"user_id":  userID,
		"status":   "awaiting_payment",
	}
	if err := s.outboxRepo.Add(ctx, tx, payload); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
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
			time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
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
	defer tx.Rollback(ctx)

	if err := s.orderRepo.SetOrderStatus(ctx, tx, orderID, entity.OrderStatusPayed); err != nil {
		return err
	}

	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"order_id": orderID,
		"user_id":  order.UserID,
		"status":   "paid",
	}
	if err := s.outboxRepo.Add(ctx, tx, payload); err != nil {
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
			time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
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
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if err := s.orderRepo.SetOrderStatus(ctx, tx, orderID, entity.OrderStatusCancelled); err != nil {
		return err
	}

	for _, item := range order.Items {
		if err := s.productRepo.ReleaseStock(ctx, tx, item.SKU, uint64(item.Count)); err != nil {
			return err
		}
	}

	payload := map[string]interface{}{
		"order_id": orderID,
		"user_id":  order.UserID,
		"status":   "cancelled",
	}
	if err := s.outboxRepo.Add(ctx, tx, payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
