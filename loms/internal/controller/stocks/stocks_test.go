package stocks

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks/mocks"
	stocksbp "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewStocksServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := mocks.NewMockstocksService(ctrl)
	require.NotNil(t, NewStocksServer(svc))
}

func TestGetStock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockstocksService(ctrl)
	svc.EXPECT().GetStock(gomock.Any(), uint32(10)).Return(uint64(50), nil)

	resp, err := NewStocksServer(svc).GetStock(context.Background(), &stocksbp.GetStockRequest{Sku: 10})
	require.NoError(t, err)
	require.Equal(t, uint64(50), resp.Count)
}

func TestGetStock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockstocksService(ctrl)
	svc.EXPECT().GetStock(gomock.Any(), uint32(10)).Return(uint64(0), errors.New("db error"))

	_, err := NewStocksServer(svc).GetStock(context.Background(), &stocksbp.GetStockRequest{Sku: 10})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.Internal, st.Code())
}

func TestSetStock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockstocksService(ctrl)
	svc.EXPECT().SetStock(gomock.Any(), uint32(10), uint64(100)).Return(nil)

	_, err := NewStocksServer(svc).SetStock(context.Background(), &stocksbp.SetStockRequest{Sku: 10, Count: 100})
	require.NoError(t, err)
}

func TestSetStock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockstocksService(ctrl)
	svc.EXPECT().SetStock(gomock.Any(), uint32(10), uint64(100)).Return(errors.New("write failed"))

	_, err := NewStocksServer(svc).SetStock(context.Background(), &stocksbp.SetStockRequest{Sku: 10, Count: 100})
	require.Error(t, err)
	st, _ := status.FromError(err)
	require.Equal(t, codes.Internal, st.Code())
}
