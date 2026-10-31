package cart

import (
	"context"

	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) CheckoutCart(ctx context.Context, req *cartpb.CheckoutCartRequest) (*cartpb.CheckoutCartResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	orderID, err := s.cartService.CheckoutCart(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}
	return &cartpb.CheckoutCartResponse{
		OrderId: orderID,
	}, nil
}
