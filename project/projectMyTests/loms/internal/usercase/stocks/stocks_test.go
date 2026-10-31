package stocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRepo struct{}

func TestNewStocksService(t *testing.T) {
	repo := &mockRepo{}
	service := NewStocksService(repo)
	require.NotNil(t, service)
	require.Equal(t, repo, service.stocksRepository)
}
