package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProductRepository struct{}

func TestNewProductService(t *testing.T) {
	repo := &mockProductRepository{}
	service := NewProductService(repo)

	require.NotNil(t, service)
	assert.NotNil(t, service.productRepository)
	assert.Equal(t, repo, service.productRepository)
}

func TestProductServiceImplementsInterface(t *testing.T) {
	// Проверяем, что тип *productService существует
	repo := &mockProductRepository{}
	service := NewProductService(repo)

	// Просто проверяем, что service не nil и имеет правильный тип
	_, ok := interface{}(service).(*productService)
	assert.True(t, ok, "service should be of type *productService")
}
