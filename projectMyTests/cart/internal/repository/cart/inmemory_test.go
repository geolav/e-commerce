package cart

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestInMemoryCartRepository(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repo *inMemoryCartRepository)
	}{
		{
			name: "AddItem new item",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				err := repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
				require.NoError(t, err)
				items, _ := repo.ListItems(context.Background(), 1)
				require.Len(t, items, 1)
				require.Equal(t, uint32(2), items[0].Count)
			},
		},
		{
			name: "AddItem accumulates count",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
				_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 3})
				items, _ := repo.ListItems(context.Background(), 1)
				require.Len(t, items, 1)
				require.Equal(t, uint32(5), items[0].Count)
			},
		},
		{
			name: "DeleteItem removes correct SKU",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
				_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 2, Count: 1})
				_ = repo.DeleteItem(context.Background(), 1, 1)
				items, _ := repo.ListItems(context.Background(), 1)
				require.Len(t, items, 1)
				require.Equal(t, uint32(2), items[0].SKU)
			},
		},
		{
			name: "DeleteItem non-existing is no-op",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				err := repo.DeleteItem(context.Background(), 1, 999)
				require.NoError(t, err)
			},
		},
		{
			name: "ClearCart empties all items",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
				_ = repo.ClearCart(context.Background(), 1)
				items, _ := repo.ListItems(context.Background(), 1)
				require.Empty(t, items)
			},
		},
		{
			name: "ListItems empty cart",
			run: func(t *testing.T, repo *inMemoryCartRepository) {
				items, err := repo.ListItems(context.Background(), 1)
				require.NoError(t, err)
				require.Empty(t, items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, NewInMemoryRepository())
		})
	}
}
