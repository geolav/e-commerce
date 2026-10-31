package cart

import (
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) ListCart(req *cartpb.ListCartRequest, stream cartpb.Cart_ListCartServer) error {
	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	items, totalPrice, err := s.cartService.ListCart(stream.Context(), req.GetUserId())
	if err != nil {
		return mapError(err)
	}

	// Если корзина пуста — ничего не отправляем
	if len(items) == 0 {
		return nil
	}

	pbItems := make([]*cartpb.Item, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, toProtoItem(item))
	}
	return stream.Send(&cartpb.ListCartResponse{
		Items:      pbItems,
		TotalPrice: totalPrice,
	})
}
