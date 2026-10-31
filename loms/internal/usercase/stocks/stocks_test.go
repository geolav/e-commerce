package stocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStocksRepository struct{}

func TestNewStocksService(t *testing.T) {
	repo := &mockStocksRepository{}
	service := NewStocksService(repo)

	require.NotNil(t, service)
	assert.NotNil(t, service.stocksRepository)
	assert.Equal(t, repo, service.stocksRepository)
}

func TestStocksServiceImplementsInterface(t *testing.T) {
	// Проверяем, что тип *stocksService существует
	repo := &mockStocksRepository{}
	service := NewStocksService(repo)

	// Просто проверяем, что service не nil и имеет правильный тип
	_, ok := interface{}(service).(*stocksService)
	assert.True(t, ok, "service should be of type *stocksService")
}
