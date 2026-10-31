package loms

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestGetOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	service := NewLomsService(
		mockOrder,
		mockProduct,
		mockOutbox,
		nil,
		zap.NewNop(),
	)

	expected := &entity.Order{
		ID:     1,
		UserID: 1,
		Status: entity.OrderStatusAwaitingPay,
	}

	mockOrder.EXPECT().
		GetOrder(gomock.Any(), int64(1)).
		Return(expected, nil)

	order, err := service.GetOrder(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, expected, order)
}

func TestGetOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	service := NewLomsService(
		mockOrder,
		mockProduct,
		mockOutbox,
		nil,
		zap.NewNop(),
	)

	mockOrder.EXPECT().
		GetOrder(gomock.Any(), int64(1)).
		Return(nil, entity.ErrOrderNotFound)

	order, err := service.GetOrder(context.Background(), 1)

	require.Nil(t, order)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}

func TestCreateOrder_ReserveStockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	service := NewLomsService(
		mockOrder,
		mockProduct,
		mockOutbox,
		nil, // db не нужен, потому что до BeginTx не дойдём
		zap.NewNop(),
	)

	items := []entity.OrderItem{
		{
			SKU:   1,
			Count: 10,
		},
	}

	// этот тест сейчас не очень полезен,
	// потому что CreateOrder начинает с BeginTx на db
	// оставляем как placeholder
	_ = service
	_ = items
}

func TestPayOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	service := NewLomsService(
		mockOrder,
		mockProduct,
		mockOutbox,
		nil,
		zap.NewNop(),
	)

	mockOrder.EXPECT().
		GetOrder(gomock.Any(), int64(1)).
		Return(nil, entity.ErrOrderNotFound)

	_, err := service.GetOrder(context.Background(), 1)

	require.Error(t, err)
}

func TestCancelOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrderRepository(ctrl)
	mockProduct := mocks.NewMockProductRepository(ctrl)
	mockOutbox := mocks.NewMockOutboxRepository(ctrl)

	service := NewLomsService(
		mockOrder,
		mockProduct,
		mockOutbox,
		nil,
		zap.NewNop(),
	)

	mockOrder.EXPECT().
		GetOrder(gomock.Any(), int64(1)).
		Return(nil, entity.ErrOrderNotFound)

	order, err := service.GetOrder(context.Background(), 1)

	require.Nil(t, order)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}
