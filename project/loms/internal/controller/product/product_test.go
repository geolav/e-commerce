package product

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewProductServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockproductRepository(ctrl)
	require.NotNil(t, NewProductServer(repo))
}

func TestCreateProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	repo.EXPECT().CreateProduct(gomock.Any(), "iphone", uint32(100)).Return(uint32(42), nil)

	resp, err := NewProductServer(repo).CreateProduct(context.Background(),
		&productpb.CreateProductRequest{Name: "iphone", Price: 100})
	require.NoError(t, err)
	require.Equal(t, uint32(42), resp.Sku)
}

func TestCreateProduct_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	repo.EXPECT().CreateProduct(gomock.Any(), "iphone", uint32(100)).Return(uint32(0), errors.New("db error"))

	_, err := NewProductServer(repo).CreateProduct(context.Background(),
		&productpb.CreateProductRequest{Name: "iphone", Price: 100})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.Internal, st.Code())
}

func TestGetProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	repo.EXPECT().GetProduct(gomock.Any(), uint32(1)).
		Return(&entity.Product{SKU: 1, Name: "iphone", Price: 100}, nil)

	resp, err := NewProductServer(repo).GetProduct(context.Background(),
		&productpb.GetProductRequest{Sku: 1})
	require.NoError(t, err)
	require.Equal(t, "iphone", resp.Name)
	require.Equal(t, uint32(100), resp.Price)
}

func TestGetProduct_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	repo.EXPECT().GetProduct(gomock.Any(), uint32(99)).Return(nil, entity.ErrProductNotFound)

	_, err := NewProductServer(repo).GetProduct(context.Background(),
		&productpb.GetProductRequest{Sku: 99})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestGetProduct_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	repo.EXPECT().GetProduct(gomock.Any(), uint32(1)).Return(nil, errors.New("db error"))

	_, err := NewProductServer(repo).GetProduct(context.Background(),
		&productpb.GetProductRequest{Sku: 1})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.Internal, st.Code())
}
