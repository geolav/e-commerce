package product

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestInMemoryProductRepository(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repo *InMemoryRepository)
	}{
		{
			name: "CreateProduct assigns SKU from 1",
			run: func(t *testing.T, repo *InMemoryRepository) {
				sku, err := repo.CreateProduct(context.Background(), "Кроссовки", 1000)
				require.NoError(t, err)
				require.Equal(t, uint32(1), sku)
			},
		},
		{
			name: "GetProduct success",
			run: func(t *testing.T, repo *InMemoryRepository) {
				sku, _ := repo.CreateProduct(context.Background(), "Кроссовки", 1000)
				p, err := repo.GetProduct(context.Background(), sku)
				require.NoError(t, err)
				require.Equal(t, "Кроссовки", p.Name)
				require.Equal(t, uint32(1000), p.Price)
			},
		},
		{
			name: "GetProduct not found",
			run: func(t *testing.T, repo *InMemoryRepository) {
				_, err := repo.GetProduct(context.Background(), 999)
				require.ErrorIs(t, err, entity.ErrProductNotFound)
			},
		},
		{
			name: "GetStock default zero",
			run: func(t *testing.T, repo *InMemoryRepository) {
				count, err := repo.GetStock(context.Background(), 1)
				require.NoError(t, err)
				require.Equal(t, uint64(0), count)
			},
		},
		{
			name: "SetStock then GetStock",
			run: func(t *testing.T, repo *InMemoryRepository) {
				err := repo.SetStock(context.Background(), 1, 100)
				require.NoError(t, err)
				count, _ := repo.GetStock(context.Background(), 1)
				require.Equal(t, uint64(100), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, NewInMemoryRepository())
		})
	}
}
