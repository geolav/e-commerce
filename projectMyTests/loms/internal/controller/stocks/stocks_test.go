package stocks

import (
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewStocksServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockstocksService(ctrl)
	server := NewStocksServer(service)
	require.NotNil(t, server)
}
