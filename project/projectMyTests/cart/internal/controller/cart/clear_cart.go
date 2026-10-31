package cart

import (
	"context"

	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cartServer) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.cartService.ClearCart(ctx, req.GetUserId()); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}
