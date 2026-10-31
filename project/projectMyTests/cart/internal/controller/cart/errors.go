package cart

import (
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, entity.ErrProductNotFound):
		return status.Errorf(codes.NotFound, "product not found")
	case errors.Is(err, entity.ErrInsufficientStock):
		return status.Errorf(codes.FailedPrecondition, "insufficient stock")
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
