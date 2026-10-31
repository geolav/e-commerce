package loms

import (
	"context"
	"testing"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	srv := NewLomsServer(&mockLomsService{})
	require.NotNil(t, srv)
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		svcErr      error
		wantOrderID int64
		wantCode    codes.Code
	}{
		{"success", nil, 42, 0},
		{"insufficient stock", entity.ErrInsufficientStock, 0, codes.FailedPrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockLomsService{
				createOrder: func(_ context.Context, _ int64, _ []entity.OrderItem) (int64, error) {
					if tt.svcErr != nil {
						return 0, tt.svcErr
					}
					return 42, nil
				},
			}
			srv := NewLomsServer(svc)
			resp, err := srv.CreateOrder(context.Background(), &pbloms.CreateOrderRequest{
				UserId: 1,
				Items:  []*pbloms.Item{{Sku: 100, Count: 2}},
			})
			if tt.wantCode != 0 {
				require.Error(t, err)
				st, _ := status.FromError(err)
				require.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantOrderID, resp.GetOrderId())
			}
		})
	}
}

func TestGetOrder(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{"success", nil, 0},
		{"not found", entity.ErrOrderNotFound, codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockLomsService{
				getOrder: func(_ context.Context, _ int64) (*entity.Order, error) {
					if tt.svcErr != nil {
						return nil, tt.svcErr
					}
					return &entity.Order{
						ID: 1, UserID: 1, Status: entity.OrderStatusAwaitingPay,
						Items:     []entity.OrderItem{{SKU: 100, Count: 2}},
						CreatedAt: now, UpdatedAt: now,
					}, nil
				},
			}
			srv := NewLomsServer(svc)
			resp, err := srv.GetOrder(context.Background(), &pbloms.GetOrderRequest{OrderId: 1})
			if tt.wantCode != 0 {
				require.Error(t, err)
				st, _ := status.FromError(err)
				require.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				require.Equal(t, int64(1), resp.GetUserId())
				require.Len(t, resp.GetItems(), 1)
			}
		})
	}
}

func TestPayOrder(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{"success", nil, 0},
		{"invalid status", entity.ErrInvalidStatus, codes.FailedPrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockLomsService{
				payOrder: func(_ context.Context, _ int64) error { return tt.svcErr },
			}
			srv := NewLomsServer(svc)
			_, err := srv.PayOrder(context.Background(), &pbloms.PayOrderRequest{OrderId: 1})
			if tt.wantCode != 0 {
				require.Error(t, err)
				st, _ := status.FromError(err)
				require.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCancelOrder(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{"success", nil, 0},
		{"invalid status", entity.ErrInvalidStatus, codes.FailedPrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockLomsService{
				cancelOrder: func(_ context.Context, _ int64) error { return tt.svcErr },
			}
			srv := NewLomsServer(svc)
			_, err := srv.CancelOrder(context.Background(), &pbloms.CancelOrderRequest{OrderId: 1})
			if tt.wantCode != 0 {
				require.Error(t, err)
				st, _ := status.FromError(err)
				require.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
