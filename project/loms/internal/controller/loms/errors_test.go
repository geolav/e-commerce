package loms

import (
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapError_AllCases(t *testing.T) {
	tests := []struct {
		err      error
		wantCode codes.Code
	}{
		{entity.ErrOrderNotFound, codes.NotFound},
		{entity.ErrInsufficientStock, codes.FailedPrecondition},
		{entity.ErrInvalidStatus, codes.FailedPrecondition},
		{errors.New("unknown error"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			err := mapError(tt.err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
