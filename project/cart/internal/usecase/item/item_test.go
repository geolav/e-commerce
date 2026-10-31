package item

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestItemService(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient)
		run       func(svc *itemService) error
		wantErr   error
	}{
		{
			name: "AddItem success",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(&entity.Product{Name: "Test", Price: 100}, nil)
				loms.EXPECT().GetStocks(gomock.Any(), uint32(1)).
					Return(uint64(10), nil)
				repo.EXPECT().AddItem(gomock.Any(), int64(1), entity.OrderItem{SKU: 1, Count: 2}).
					Return(nil)
			},
			run: func(svc *itemService) error {
				return svc.AddItem(context.Background(), 1, 1, 2)
			},
		},
		{
			name: "AddItem product not found",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(nil, entity.ErrProductNotFound)
			},
			run: func(svc *itemService) error {
				return svc.AddItem(context.Background(), 1, 1, 2)
			},
			wantErr: entity.ErrProductNotFound,
		},
		{
			name: "AddItem insufficient stock",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				product.EXPECT().GetProductInfo(gomock.Any(), uint32(1)).
					Return(&entity.Product{Name: "Test", Price: 100}, nil)
				loms.EXPECT().GetStocks(gomock.Any(), uint32(1)).
					Return(uint64(1), nil)
			},
			run: func(svc *itemService) error {
				return svc.AddItem(context.Background(), 1, 1, 5)
			},
			wantErr: entity.ErrInsufficientStock,
		},
		{
			name: "DeleteItem success",
			mockSetup: func(repo *mocks.MockcartRepository, product *mocks.MockproductClient, loms *mocks.MocklomsClient) {
				repo.EXPECT().DeleteItem(gomock.Any(), int64(1), uint32(1)).Return(nil)
			},
			run: func(svc *itemService) error {
				return svc.DeleteItem(context.Background(), 1, 1)
			},
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

			svc := NewItemService(repo, product, loms)
			err := tt.run(svc)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
