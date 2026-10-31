package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
)

func TestListCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewCartService(mockRepo, mockProduct, mockLoms)

	mockRepo.EXPECT().
		ListItems(gomock.Any(), int64(1)).
		Return([]entity.OrderItem{{SKU: 1, Count: 2}}, nil)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)

	items, totalPrice, err := service.ListCart(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint32(2000), totalPrice)
	require.Equal(t, "Кроссовки", items[0].Name)
}

func TestListCart_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewCartService(mockRepo, mockProduct, mockLoms)

	mockRepo.EXPECT().
		ListItems(gomock.Any(), int64(1)).
		Return(nil, errors.New("repo error"))

	_, _, err := service.ListCart(context.Background(), 1)
	require.Error(t, err)
}

func TestClearCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewCartService(mockRepo, mockProduct, mockLoms)

	mockRepo.EXPECT().
		ClearCart(gomock.Any(), int64(1)).
		Return(nil)

	err := service.ClearCart(context.Background(), 1)
	require.NoError(t, err)
}

func TestCheckoutCart_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewCartService(mockRepo, mockProduct, mockLoms)

	items := []entity.OrderItem{{SKU: 1, Count: 2}}

	mockRepo.EXPECT().
		ListItems(gomock.Any(), int64(1)).
		Return(items, nil)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)

	mockLoms.EXPECT().
		CreateOrder(gomock.Any(), int64(1), gomock.Any()).
		Return(int64(42), nil)

	mockRepo.EXPECT().
		ClearCart(gomock.Any(), int64(1)).
		Return(nil)

	orderID, err := service.CheckoutCart(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(42), orderID)
}

func TestCheckoutCart_CreateOrderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcartRepository(ctrl)
	mockProduct := mocks.NewMockproductClient(ctrl)
	mockLoms := mocks.NewMocklomsClient(ctrl)

	service := NewCartService(mockRepo, mockProduct, mockLoms)

	items := []entity.OrderItem{{SKU: 1, Count: 2}}

	mockRepo.EXPECT().
		ListItems(gomock.Any(), int64(1)).
		Return(items, nil)

	mockProduct.EXPECT().
		GetProductInfo(gomock.Any(), uint32(1)).
		Return(&entity.Product{Name: "Кроссовки", Price: 1000}, nil)

	mockLoms.EXPECT().
		CreateOrder(gomock.Any(), int64(1), gomock.Any()).
		Return(int64(0), errors.New("loms error"))

	_, err := service.CheckoutCart(context.Background(), 1)
	require.Error(t, err)
}
