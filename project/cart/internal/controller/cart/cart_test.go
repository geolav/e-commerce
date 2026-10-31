package cart

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddItem(t *testing.T) {
	tests := []struct {
		name     string
		req      *cartpb.AddItemRequest
		svcErr   error
		wantCode codes.Code
	}{
		{
			name: "success",
			req:  &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
		},
		{
			name:     "product not found",
			req:      &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
			svcErr:   entity.ErrProductNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "insufficient stock",
			req:      &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
			svcErr:   entity.ErrInsufficientStock,
			wantCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItem := mocks.NewMockItemService(ctrl)
			mockCart := mocks.NewMockService(ctrl)
			srv := NewCartServer(mockItem, mockCart)

			mockItem.EXPECT().
				AddItem(gomock.Any(), tt.req.GetUserId(), tt.req.GetSku(), tt.req.GetCount()).
				Return(tt.svcErr)

			_, err := srv.AddItem(context.Background(), tt.req)

			if tt.wantCode != 0 {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteItem(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		sku    uint32
		svcErr error
	}{
		{"success", 1, 1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItem := mocks.NewMockItemService(ctrl)
			mockCart := mocks.NewMockService(ctrl)
			srv := NewCartServer(mockItem, mockCart)

			mockItem.EXPECT().
				DeleteItem(gomock.Any(), tt.userID, tt.sku).
				Return(tt.svcErr)

			_, err := srv.DeleteItem(context.Background(), &cartpb.DeleteItemRequest{UserId: tt.userID, Sku: tt.sku})
			require.NoError(t, err)
		})
	}
}

func TestClearCart(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		svcErr error
	}{
		{"success", 1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItem := mocks.NewMockItemService(ctrl)
			mockCart := mocks.NewMockService(ctrl)
			srv := NewCartServer(mockItem, mockCart)

			mockCart.EXPECT().ClearCart(gomock.Any(), tt.userID).Return(tt.svcErr)

			_, err := srv.ClearCart(context.Background(), &cartpb.ClearCartRequest{UserId: tt.userID})
			require.NoError(t, err)
		})
	}
}

func TestCheckoutCart(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		wantID  int64
		svcErr  error
		wantErr bool
	}{
		{"success", 1, 42, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItem := mocks.NewMockItemService(ctrl)
			mockCart := mocks.NewMockService(ctrl)
			srv := NewCartServer(mockItem, mockCart)

			mockCart.EXPECT().CheckoutCart(gomock.Any(), tt.userID).Return(tt.wantID, tt.svcErr)

			_, err := srv.CheckoutCart(context.Background(), &cartpb.CheckoutCartRequest{UserId: tt.userID})
			require.NoError(t, err)
		})
	}
}

func TestListCart(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		items   []entity.OrderItem
		wantLen int
	}{
		{
			name:    "success",
			userID:  1,
			items:   []entity.OrderItem{{SKU: 1, Count: 2, Name: "Кроссовки", Price: 1000}},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItem := mocks.NewMockItemService(ctrl)
			mockCart := mocks.NewMockService(ctrl)
			srv := NewCartServer(mockItem, mockCart)

			mockCart.EXPECT().
				ListCart(gomock.Any(), tt.userID).
				Return(tt.items, uint32(2000), nil)

			stream := &mockListCartStream{ctx: context.Background()}
			err := srv.ListCart(&cartpb.ListCartRequest{UserId: tt.userID}, stream)
			require.NoError(t, err)
			require.Len(t, stream.responses, tt.wantLen)
		})
	}
}

type mockListCartStream struct {
	cartpb.Cart_ListCartServer
	ctx       context.Context
	responses []*cartpb.ListCartResponse
}

func (m *mockListCartStream) Send(r *cartpb.ListCartResponse) error {
	m.responses = append(m.responses, r)
	return nil
}

func (m *mockListCartStream) Context() context.Context { return m.ctx }
