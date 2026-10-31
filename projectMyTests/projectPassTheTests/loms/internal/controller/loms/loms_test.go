package loms

import (
	"context"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockLomsService реализует интерфейс lomsService для тестов
type mockLomsService struct {
	createOrder func(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error)
	getOrder    func(ctx context.Context, orderID int64) (*entity.Order, error)
	payOrder    func(ctx context.Context, orderID int64) error
	cancelOrder func(ctx context.Context, orderID int64) error
}

func (m *mockLomsService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	if m.createOrder != nil {
		return m.createOrder(ctx, userID, items)
	}
	return 0, nil
}

func (m *mockLomsService) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	if m.getOrder != nil {
		return m.getOrder(ctx, orderID)
	}
	return nil, nil
}

func (m *mockLomsService) PayOrder(ctx context.Context, orderID int64) error {
	if m.payOrder != nil {
		return m.payOrder(ctx, orderID)
	}
	return nil
}

func (m *mockLomsService) CancelOrder(ctx context.Context, orderID int64) error {
	if m.cancelOrder != nil {
		return m.cancelOrder(ctx, orderID)
	}
	return nil
}

func TestNewLomsServer(t *testing.T) {
	svc := &mockLomsService{}
	srv := NewLomsServer(svc)
	require.NotNil(t, srv)
}

func TestCreateOrder(t *testing.T) {
	svc := &mockLomsService{
		createOrder: func(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
			return 42, nil
		},
	}
	srv := NewLomsServer(svc)

	resp, err := srv.CreateOrder(context.Background(), &pbloms.CreateOrderRequest{
		UserId: 1,
		Items:  []*pbloms.Item{{Sku: 100, Count: 2}},
	})
	require.NoError(t, err)
	require.Equal(t, int64(42), resp.GetOrderId())
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	svc := &mockLomsService{
		createOrder: func(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
			return 0, lomsusecase.ErrInsufficientStock
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.CreateOrder(context.Background(), &pbloms.CreateOrderRequest{
		UserId: 1,
		Items:  []*pbloms.Item{{Sku: 100, Count: 999}},
	})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestGetOrder(t *testing.T) {
	now := time.Now()
	svc := &mockLomsService{
		getOrder: func(ctx context.Context, orderID int64) (*entity.Order, error) {
			return &entity.Order{
				ID:        1,
				UserID:    1,
				Status:    entity.OrderStatusAwaitingPay,
				Items:     []entity.OrderItem{{SKU: 100, Count: 2}},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}
	srv := NewLomsServer(svc)

	resp, err := srv.GetOrder(context.Background(), &pbloms.GetOrderRequest{OrderId: 1})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.GetUserId())
	require.Len(t, resp.GetItems(), 1)
	require.Equal(t, uint32(100), resp.GetItems()[0].GetSku())
	require.NotNil(t, resp.GetCreatedAt())
}

func TestGetOrder_NotFound(t *testing.T) {
	svc := &mockLomsService{
		getOrder: func(ctx context.Context, orderID int64) (*entity.Order, error) {
			return nil, lomsusecase.ErrOrderNotFound
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.GetOrder(context.Background(), &pbloms.GetOrderRequest{OrderId: 999})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestPayOrder(t *testing.T) {
	svc := &mockLomsService{
		payOrder: func(ctx context.Context, orderID int64) error {
			return nil
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.PayOrder(context.Background(), &pbloms.PayOrderRequest{OrderId: 1})
	require.NoError(t, err)
}

func TestPayOrder_InvalidStatus(t *testing.T) {
	svc := &mockLomsService{
		payOrder: func(ctx context.Context, orderID int64) error {
			return lomsusecase.ErrInvalidStatus
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.PayOrder(context.Background(), &pbloms.PayOrderRequest{OrderId: 1})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestCancelOrder(t *testing.T) {
	svc := &mockLomsService{
		cancelOrder: func(ctx context.Context, orderID int64) error {
			return nil
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.CancelOrder(context.Background(), &pbloms.CancelOrderRequest{OrderId: 1})
	require.NoError(t, err)
}

func TestCancelOrder_InvalidStatus(t *testing.T) {
	svc := &mockLomsService{
		cancelOrder: func(ctx context.Context, orderID int64) error {
			return lomsusecase.ErrInvalidStatus
		},
	}
	srv := NewLomsServer(svc)

	_, err := srv.CancelOrder(context.Background(), &pbloms.CancelOrderRequest{OrderId: 1})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}
