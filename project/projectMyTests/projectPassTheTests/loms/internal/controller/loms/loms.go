package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	lomsService interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) // TODO mb change to []*entity.OrderItem
		GetOrder(ctx context.Context, orderID int64) (*entity.Order, error)
		PayOrder(ctx context.Context, orderID int64) error
		CancelOrder(ctx context.Context, orderID int64) error
	}
)

type lomsServer struct {
	lomsService lomsService
}

func NewLomsServer(lomsService lomsService) *lomsServer {
	return &lomsServer{
		lomsService: lomsService,
	}
}

func (s *lomsServer) CreateOrder(ctx context.Context, req *pbloms.CreateOrderRequest) (*pbloms.CreateOrderResponse, error) {
	items := make([]entity.OrderItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, entity.OrderItem{
			SKU:   item.GetSku(),
			Count: item.GetCount(),
		})
	}
	orderID, err := s.lomsService.CreateOrder(ctx, req.GetUserId(), items)
	// ErrInsufficientStock — это ожидаемая бизнес-ошибка (не хватает товара на складе).
	// Клиент должен об этом знать, поэтому возвращаем FailedPrecondition — "запрос правильный, но условия не выполнены"
	// любая другая ошибка — это неожиданная внутренняя ошибка (что-то сломалось), возвращаем Internal — "что-то пошло не так на сервере"
	if err != nil {
		return nil, mapError(err)
	}
	return &pbloms.CreateOrderResponse{OrderId: orderID}, nil
}

func (s *lomsServer) GetOrder(ctx context.Context, req *pbloms.GetOrderRequest) (*pbloms.GetOrderResponse, error) {
	order, err := s.lomsService.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*pbloms.Item, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &pbloms.Item{
			Sku:   item.SKU,
			Count: item.Count,
		})
	}
	return &pbloms.GetOrderResponse{
		Items:     items,
		Status:    converter.FromOrderStatus(order.Status),
		UserId:    order.UserID,
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}, nil
}

func (s *lomsServer) PayOrder(ctx context.Context, req *pbloms.PayOrderRequest) (*emptypb.Empty, error) {
	err := s.lomsService.PayOrder(ctx, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *lomsServer) CancelOrder(ctx context.Context, req *pbloms.CancelOrderRequest) (*emptypb.Empty, error) {
	err := s.lomsService.CancelOrder(ctx, req.GetOrderId())
	if err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}
