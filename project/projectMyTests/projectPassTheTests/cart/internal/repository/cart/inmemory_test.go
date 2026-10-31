package cart

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

func TestAddItem_NewItem(t *testing.T) {
	repo := NewInMemoryRepository()
	err := repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
	require.NoError(t, err)

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint32(2), items[0].Count)
}

func TestAddItem_ExistingItem(t *testing.T) {
	repo := NewInMemoryRepository()
	_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
	_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 3})

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint32(5), items[0].Count)
}

func TestDeleteItem(t *testing.T) {
	repo := NewInMemoryRepository()
	_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})
	_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 2, Count: 1})

	err := repo.DeleteItem(context.Background(), 1, 1)
	require.NoError(t, err)

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, uint32(2), items[0].SKU)
}

func TestDeleteItem_NotExisting(t *testing.T) {
	repo := NewInMemoryRepository()
	err := repo.DeleteItem(context.Background(), 1, 999)
	require.NoError(t, err)
}

func TestClearCart(t *testing.T) {
	repo := NewInMemoryRepository()
	_ = repo.AddItem(context.Background(), 1, entity.OrderItem{SKU: 1, Count: 2})

	err := repo.ClearCart(context.Background(), 1)
	require.NoError(t, err)

	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestListItems_Empty(t *testing.T) {
	repo := NewInMemoryRepository()
	items, err := repo.ListItems(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, items)
}
