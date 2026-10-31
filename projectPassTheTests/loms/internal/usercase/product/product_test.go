package product

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRepo struct{}

func TestNewProductService(t *testing.T) {
	repo := &mockRepo{}
	service := NewProductService(repo)
	require.NotNil(t, service)
	require.Equal(t, repo, service.productRepository)
}
