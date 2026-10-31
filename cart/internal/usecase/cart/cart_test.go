package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartService(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient)
		run       func(svc *cartService) error
		wantErr   bool
	}{
		{
			name: "ListCart success",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().ListItems(gomock.Any(), int64(1)).
					Return([]entity.OrderItem{{SKU: 1, Count: 2}}, nil)
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)
			},
			run: func(svc *cartService) error {
				items, price, err := svc.ListCart(context.Background(), 1)
				require.NoError(t, err)
				require.Len(t, items, 1)
				require.Equal(t, uint32(2000), price)
				return nil
			},
		},
		{
			name: "ListCart repo error",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().ListItems(gomock.Any(), int64(1)).
					Return(nil, errors.New("repo error"))
			},
			run: func(svc *cartService) error {
				_, _, err := svc.ListCart(context.Background(), 1)
				return err
			},
			wantErr: true,
		},
		{
			name: "ClearCart success",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().ClearCart(gomock.Any(), int64(1)).Return(nil)
			},
			run: func(svc *cartService) error {
				return svc.ClearCart(context.Background(), 1)
			},
		},
		{
			name: "CheckoutCart success",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().ListItems(gomock.Any(), int64(1)).
					Return([]entity.OrderItem{{SKU: 1, Count: 2}}, nil)
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)
				loms.EXPECT().CreateOrder(gomock.Any(), int64(1), gomock.Any()).
					Return(int64(42), nil)
				repo.EXPECT().ClearCart(gomock.Any(), int64(1)).Return(nil)
			},
			run: func(svc *cartService) error {
				id, err := svc.CheckoutCart(context.Background(), 1)
				require.Equal(t, int64(42), id)
				return err
			},
		},
		{
			name: "CheckoutCart loms error",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().ListItems(gomock.Any(), int64(1)).
					Return([]entity.OrderItem{{SKU: 1, Count: 2}}, nil)
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)
				loms.EXPECT().CreateOrder(gomock.Any(), int64(1), gomock.Any()).
					Return(int64(0), errors.New("loms error"))
			},
			run: func(svc *cartService) error {
				_, err := svc.CheckoutCart(context.Background(), 1)
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockcartRepository(ctrl)
			product := mocks.NewMockproductClient(ctrl)
			loms := mocks.NewMocklomsClient(ctrl)
			tt.mockSetup(repo, product, loms)

			svc := NewCartService(repo, product, loms)
			err := tt.run(svc)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
