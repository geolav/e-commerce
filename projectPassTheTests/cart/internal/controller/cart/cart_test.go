package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"testing"

	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

func TestAddItem(t *testing.T) {
	tests := []struct {
		name    string
		req     *cartpb.AddItemRequest
		mockErr error
		wantErr codes.Code
	}{
		{
			name:    "success",
			req:     &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
			mockErr: nil,
		},
		{
			name:    "product not found",
			req:     &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
			mockErr: entity.ErrProductNotFound,
			wantErr: codes.NotFound,
		},
		{
			name:    "insufficient stock",
			req:     &cartpb.AddItemRequest{UserId: 1, Sku: 1, Count: 2},
			mockErr: entity.ErrInsufficientStock,
			wantErr: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItemService := mocks.NewMockItemService(ctrl)
			mockCartService := mocks.NewMockService(ctrl)
			server := NewCartServer(mockItemService, mockCartService)

			mockItemService.EXPECT().
				AddItem(gomock.Any(), tt.req.GetUserId(), tt.req.GetSku(), tt.req.GetCount()).
				Return(tt.mockErr)

			_, err := server.AddItem(context.Background(), tt.req)

			if tt.wantErr == 0 {
				require.NoError(t, err)
			} else {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantErr, st.Code())
			}
		})
	}
}

//func TestAddItem_Success(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockItemService := mocks.NewMockItemService(ctrl)
//	mockCartService := mocks.NewMockService(ctrl)
//	server := NewCartServer(mockItemService, mockCartService)
//
//	mockItemService.EXPECT().
//		AddItem(gomock.Any(), int64(1), uint32(1), uint32(2)).
//		Return(nil)
//
//	_, err := server.AddItem(context.Background(), &cartpb.AddItemRequest{
//		UserId: 1,
//		Sku:    1,
//		Count:  2,
//	})
//
//	require.NoError(t, err)
//}
//
//func TestAddItem_ProductNotFound(t *testing.T) {
//	controller := gomock.NewController(t)
//	defer controller.Finish()
//	mockItemService := mocks.NewMockItemService(controller)
//	mockCartService := mocks.NewMockService(controller)
//	server := NewCartServer(mockItemService, mockCartService)
//
//	mockItemService.EXPECT().
//		AddItem(gomock.Any(), int64(1), uint32(1), uint32(2)).
//		Return(entity.ErrProductNotFound)
//
//	_, err := server.AddItem(context.Background(), &cartpb.AddItemRequest{
//		UserId: 1,
//		Sku:    1,
//		Count:  2,
//	})
//
//	require.Error(t, err)
//	st, ok := status.FromError(err)
//	require.True(t, ok)
//	require.Equal(t, codes.NotFound, st.Code())
//}
//
//func TestAddItem_InsufficientStock(t *testing.T) {
//	controller := gomock.NewController(t)
//	defer controller.Finish()
//	mockItemService := mocks.NewMockItemService(controller)
//	mockCartService := mocks.NewMockService(controller)
//	server := NewCartServer(mockItemService, mockCartService)
//
//	mockItemService.EXPECT().
//		AddItem(gomock.Any(), int64(1), uint32(1), uint32(2)).
//		Return(entity.ErrInsufficientStock)
//
//	_, err := server.AddItem(context.Background(), &cartpb.AddItemRequest{
//		UserId: 1,
//		Sku:    1,
//		Count:  2,
//	})
//
//	require.Error(t, err)
//	st, ok := status.FromError(err)
//	require.True(t, ok)
//	require.Equal(t, codes.FailedPrecondition, st.Code())
//}

func TestDeleteItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemService := mocks.NewMockItemService(ctrl)
	mockCartService := mocks.NewMockService(ctrl)
	server := NewCartServer(mockItemService, mockCartService)

	mockItemService.EXPECT().
		DeleteItem(gomock.Any(), int64(1), uint32(1)).
		Return(nil)

	_, err := server.DeleteItem(context.Background(), &cartpb.DeleteItemRequest{
		UserId: 1,
		Sku:    1,
	})

	require.NoError(t, err)
}

func TestClearCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemService := mocks.NewMockItemService(ctrl)
	mockCartService := mocks.NewMockService(ctrl)
	server := NewCartServer(mockItemService, mockCartService)

	mockCartService.EXPECT().
		ClearCart(gomock.Any(), int64(1)).
		Return(nil)

	_, err := server.ClearCart(context.Background(), &cartpb.ClearCartRequest{
		UserId: 1,
	})

	require.NoError(t, err)
}

func TestCheckoutCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemService := mocks.NewMockItemService(ctrl)
	mockCartService := mocks.NewMockService(ctrl)
	server := NewCartServer(mockItemService, mockCartService)

	mockCartService.EXPECT().
		CheckoutCart(gomock.Any(), int64(1)).
		Return(int64(42), nil)

	_, err := server.CheckoutCart(context.Background(), &cartpb.CheckoutCartRequest{
		UserId: 1,
	})

	require.NoError(t, err)
}

func TestListCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockItemService := mocks.NewMockItemService(ctrl)
	mockCartService := mocks.NewMockService(ctrl)
	server := NewCartServer(mockItemService, mockCartService)

	items := []entity.OrderItem{{SKU: 1, Count: 2, Name: "Кроссовки", Price: 1000}}
	mockCartService.EXPECT().
		ListCart(gomock.Any(), int64(1)).
		Return(items, uint32(2000), nil)

	stream := &mockListCartStream{ctx: context.Background()}
	err := server.ListCart(&cartpb.ListCartRequest{UserId: 1}, stream)
	require.NoError(t, err)
	require.Len(t, stream.responses, 1)
}

type mockListCartStream struct {
	cartpb.Cart_ListCartServer
	ctx       context.Context
	responses []*cartpb.ListCartResponse
}

func (m *mockListCartStream) Send(resp *cartpb.ListCartResponse) error {
	m.responses = append(m.responses, resp)
	return nil
}

func (m *mockListCartStream) Context() context.Context {
	return m.ctx
}
