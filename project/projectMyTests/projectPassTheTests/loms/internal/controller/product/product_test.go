package product

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewProductServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockproductRepository(ctrl)
	server := NewProductServer(repo)
	require.NotNil(t, server)
}
