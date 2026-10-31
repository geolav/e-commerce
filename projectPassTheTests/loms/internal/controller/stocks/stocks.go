package stocks

import (
	"context"

	stocksbp "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source=stocks.go -destination=mocks/stocks_mock.go -package=mocks
type (
	stocksService interface {
		GetStock(ctx context.Context, sku uint32) (uint64, error)
		SetStock(ctx context.Context, sku uint32, count uint64) error
	}
)

type stocksServer struct {
	stocksService stocksService
}

func NewStocksServer(stocksService stocksService) *stocksServer {
	return &stocksServer{
		stocksService: stocksService,
	}
}

func (s *stocksServer) GetStock(ctx context.Context, req *stocksbp.GetStockRequest) (*stocksbp.GetStockResponse, error) {
	count, err := s.stocksService.GetStock(ctx, req.GetSku())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting stock: %v", err)
	}
	return &stocksbp.GetStockResponse{Count: count}, nil
}

func (s *stocksServer) SetStock(ctx context.Context, req *stocksbp.SetStockRequest) (*emptypb.Empty, error) {
	err := s.stocksService.SetStock(ctx, req.GetSku(), req.GetCount())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error setting stock: %v", err)
	}
	return &emptypb.Empty{}, nil
}
