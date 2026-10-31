package product

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	repo := NewInMemoryRepository()

	sku, err := repo.CreateProduct(context.Background(), "Кроссовки", 1000)
	require.NoError(t, err)
	require.Equal(t, uint32(1), sku)
}

func TestGetProduct_Success(t *testing.T) {
	repo := NewInMemoryRepository()
	sku, _ := repo.CreateProduct(context.Background(), "Кроссовки", 1000)

	product, err := repo.GetProduct(context.Background(), sku)
	require.NoError(t, err)
	require.Equal(t, "Кроссовки", product.Name)
	require.Equal(t, uint32(1000), product.Price)
}

func TestGetProduct_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	_, err := repo.GetProduct(context.Background(), 999)
	require.ErrorIs(t, err, entity.ErrProductNotFound)
}

func TestGetStock(t *testing.T) {
	repo := NewInMemoryRepository()

	count, err := repo.GetStock(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
}

func TestSetStock(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SetStock(context.Background(), 1, 100)
	require.NoError(t, err)

	count, _ := repo.GetStock(context.Background(), 1)
	require.Equal(t, uint64(100), count)
}
