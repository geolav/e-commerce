package loms

import (
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, entity.ErrOrderNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, entity.ErrInvalidStatus):
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	case errors.Is(err, entity.ErrInsufficientStock):
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
