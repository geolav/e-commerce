package loms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms/mocks"
)

func TestCreateOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	items := []entity.OrderItem{{SKU: 1, Count: 2}}

	mockStocks.EXPECT().ReserveStock(gomock.Any(), uint32(1), uint64(2)).Return(nil)
	mockOrder.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	orderID, err := service.CreateOrder(context.Background(), 1, items)
	require.NoError(t, err)
	require.Equal(t, int64(1), orderID)
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	items := []entity.OrderItem{{SKU: 1, Count: 10}}

	mockStocks.EXPECT().ReserveStock(gomock.Any(), uint32(1), uint64(10)).Return(entity.ErrInsufficientStock)
	_, err := service.CreateOrder(context.Background(), 1, items)
	require.ErrorIs(t, err, ErrInsufficientStock)
}

func TestGetOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	expected := &entity.Order{ID: 1, UserID: 1, Status: entity.OrderStatusAwaitingPay}
	mockOrder.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(expected, nil)

	order, err := service.GetOrder(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, expected, order)
}

func TestGetOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	mockOrder.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(nil, ErrOrderNotFound)

	_, err := service.GetOrder(context.Background(), 1)
	require.Error(t, err)
}

func TestPayOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	//mockOrder.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(
	//	&entity.Order{ID: 1, Status: entity.OrderStatusAwaitingPay}, nil,
	//)
	mockOrder.EXPECT().SetOrderStatus(gomock.Any(), int64(1), entity.OrderStatusPayed).Return(nil)

	err := service.PayOrder(context.Background(), 1)
	require.NoError(t, err)
}

func TestPayOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	mockOrder.EXPECT().SetOrderStatus(gomock.Any(), int64(1), entity.OrderStatusPayed).Return(ErrOrderNotFound)

	err := service.PayOrder(context.Background(), 1)
	require.ErrorIs(t, err, ErrOrderNotFound)
}

func TestPayOrder_InvalidStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	mockOrder.EXPECT().SetOrderStatus(gomock.Any(), int64(1), entity.OrderStatusPayed).Return(ErrInvalidStatus)

	err := service.PayOrder(context.Background(), 1)
	require.ErrorIs(t, err, ErrInvalidStatus)
}

func TestCancelOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	mockOrder.EXPECT().SetOrderStatus(gomock.Any(), int64(1), entity.OrderStatusCancelled).Return(nil)
	mockOrder.EXPECT().GetOrder(gomock.Any(), int64(1)).Return(
		&entity.Order{ID: 1, Status: entity.OrderStatusCancelled, Items: []entity.OrderItem{}}, nil,
	)

	err := service.CancelOrder(context.Background(), 1)
	require.NoError(t, err)
}

func TestCancelOrder_InvalidStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockorderRepository(ctrl)
	mockStocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(mockOrder, mockStocks)

	mockOrder.EXPECT().SetOrderStatus(gomock.Any(), int64(1), entity.OrderStatusCancelled).Return(ErrInvalidStatus)

	err := service.CancelOrder(context.Background(), 1)
	require.ErrorIs(t, err, ErrInvalidStatus)
}
