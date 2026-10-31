package item

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddItem_Success(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := mocks.NewMockcartRepository(controller)
	mockProduct := mocks.NewMockproductClient(controller)
	mockLoms := mocks.NewMocklomsClient(controller)

	service := NewItemService(mockRepo, mockProduct, mockLoms)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(&entity.Product{Name: "Test", Price: 100}, nil)

	mockLoms.EXPECT().
		GetStocks(gomock.Any(), uint32(1)).
		Return(uint64(10), nil)

	mockRepo.EXPECT().
		AddItem(gomock.Any(), int64(1), entity.OrderItem{SKU: 1, Count: 2}).
		Return(nil)

	err := service.AddItem(context.Background(), 1, 1, 2)

	require.NoError(t, err)
}

func TestAddItem_ProductNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := mocks.NewMockcartRepository(controller)
	mockProduct := mocks.NewMockproductClient(controller)
	mockLoms := mocks.NewMocklomsClient(controller)

	service := NewItemService(mockRepo, mockProduct, mockLoms)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(nil, entity.ErrProductNotFound)

	err := service.AddItem(context.Background(), 1, 1, 2)

	require.ErrorIs(t, err, entity.ErrProductNotFound)
}

func TestAddItem_InsufficientStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewItemService(mockRepo, mockProduct, mockLoms)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(&entity.Product{Name: "Test", Price: 100}, nil)

	mockLoms.EXPECT().
		GetStocks(gomock.Any(), uint32(1)).
		Return(uint64(1), nil)

	err := service.AddItem(context.Background(), 1, 1, 5)
	require.ErrorIs(t, err, entity.ErrInsufficientStock)
}

func TestDeleteItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewItemService(mockRepo, mockProduct, mockLoms)

	mockRepo.EXPECT().
		DeleteItem(gomock.Any(), int64(1), uint32(1)).
		Return(nil)

	err := service.DeleteItem(context.Background(), 1, 1)
	require.NoError(t, err)
}
