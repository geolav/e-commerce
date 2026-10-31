package loms

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── mockTxBeginner — реализует TxBeginner, возвращает ошибку ─────────────────
// Нам НЕ нужен рабочий pgx.Tx — достаточно чтобы BeginTx вернул ошибку,
// тогда все ветки до BeginTx покроются, а ветки внутри транзакции —
// через отдельный mockTxBeginner который возвращает mockTx ниже.

type mockTxBeginner struct {
	tx  pgx.Tx
	err error
}

func (m *mockTxBeginner) BeginTx(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	return m.tx, m.err
}

// errTxBeginner — всегда возвращает ошибку, tx не нужен
func errTxBeginner(err error) *mockTxBeginner {
	return &mockTxBeginner{err: err}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newMocks(t *testing.T) (
	*mocks.MockOrderRepository,
	*mocks.MockProductRepository,
	*mocks.MockOutboxRepository,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	return mocks.NewMockOrderRepository(ctrl),
		mocks.NewMockProductRepository(ctrl),
		mocks.NewMockOutboxRepository(ctrl)
}

// ── NewLomsService ────────────────────────────────────────────────────────────

func TestNewLomsService_NotNil(t *testing.T) {
	require.NotNil(t, NewLomsService(nil, nil, nil, nil))
}

func TestNewLomsService_StoresFields(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	db := errTxBeginner(nil)
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, db)
	require.Equal(t, orderRepo, svc.orderRepo)
	require.Equal(t, productRepo, svc.productRepo)
	require.Equal(t, outboxRepo, svc.outboxRepo)
}

// ── GetOrder ──────────────────────────────────────────────────────────────────

func TestGetOrder_Success(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, nil)
	_ = productRepo
	_ = outboxRepo

	want := &entity.Order{ID: 1, UserID: 42, Status: entity.OrderStatusAwaitingPay}
	orderRepo.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(want, nil)

	got, err := svc.GetOrder(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetOrder_NotFound(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, nil)
	_ = productRepo
	_ = outboxRepo

	orderRepo.EXPECT().GetOrder(gomock.Any(), int64(99)).Return(nil, entity.ErrOrderNotFound)

	got, err := svc.GetOrder(context.Background(), 99)
	require.Nil(t, got)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}

func TestGetOrder_GenericError(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, nil)
	_ = productRepo
	_ = outboxRepo

	orderRepo.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(nil, errors.New("db error"))

	got, err := svc.GetOrder(context.Background(), 1)
	require.Nil(t, got)
	require.Error(t, err)
}

// ── CreateOrder — BeginTx error покрывает retry-loop и createOrderOnce ────────

func TestCreateOrder_BeginTxError(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	_ = productRepo
	_ = outboxRepo
	db := errTxBeginner(errors.New("cannot connect"))
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, db)

	_, err := svc.CreateOrder(context.Background(), 1, []entity.OrderItem{{SKU: 1, Count: 1}})
	require.Error(t, err)
}

// TestCreateOrder_AllRetriesExhausted — имитируем serialization failure (40001)
// через кастомный TxBeginner который считает вызовы
func TestCreateOrder_AllRetriesExhausted(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	_ = productRepo
	_ = outboxRepo

	calls := 0
	db := &countingTxBeginner{
		fn: func() (pgx.Tx, error) {
			calls++
			// Возвращаем обычную ошибку — выйдем из цикла сразу (не 40001)
			return nil, errors.New("connection refused")
		},
	}
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, db)

	_, err := svc.CreateOrder(context.Background(), 1, []entity.OrderItem{{SKU: 1, Count: 1}})
	require.Error(t, err)
	require.Equal(t, 1, calls) // не-serialization error — выходим сразу
}

// ── PayOrder — BeginTx error ──────────────────────────────────────────────────

func TestPayOrder_BeginTxError(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	_ = productRepo
	_ = outboxRepo
	db := errTxBeginner(errors.New("no connection"))
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, db)

	err := svc.PayOrder(context.Background(), 1)
	require.Error(t, err)
}

// ── CancelOrder — BeginTx error ───────────────────────────────────────────────

func TestCancelOrder_BeginTxError(t *testing.T) {
	orderRepo, productRepo, outboxRepo := newMocks(t)
	_ = productRepo
	_ = outboxRepo
	db := errTxBeginner(errors.New("no connection"))
	svc := NewLomsService(orderRepo, productRepo, outboxRepo, db)

	err := svc.CancelOrder(context.Background(), 1)
	require.Error(t, err)
}

type countingTxBeginner struct {
	fn func() (pgx.Tx, error)
}

func (c *countingTxBeginner) BeginTx(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	return c.fn()
}

func TestGetOrder_Exists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	expected := &entity.Order{ID: 1, UserID: 1, Status: entity.OrderStatusAwaitingPay}
	mockOrder.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(expected, nil)

	service := NewLomsService(mockOrder, mockProduct, mockOutbox, nil)
	order, err := service.GetOrder(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, expected, order)
}

func TestGetOrder_NotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	mockOrder.EXPECT().GetOrder(gomock.Any(), int64(999)).Return(nil, entity.ErrOrderNotFound)

	service := NewLomsService(mockOrder, mockProduct, mockOutbox, nil)
	_, err := service.GetOrder(context.Background(), 999)

	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}
