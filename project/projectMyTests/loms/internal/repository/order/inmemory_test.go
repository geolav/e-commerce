package order

import (
	"context"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestInMemoryOrderRepository(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repo *InMemoryRepository)
	}{
		{
			name: "CreateOrder assigns sequential ID",
			run: func(t *testing.T, repo *InMemoryRepository) {
				id, err := repo.CreateOrder(context.Background(), &entity.Order{UserID: 1})
				require.NoError(t, err)
				require.Equal(t, int64(1), id)

				id2, err := repo.CreateOrder(context.Background(), &entity.Order{UserID: 2})
				require.NoError(t, err)
				require.Equal(t, int64(2), id2)
			},
		},
		{
			name: "GetOrder success",
			run: func(t *testing.T, repo *InMemoryRepository) {
				id, _ := repo.CreateOrder(context.Background(), &entity.Order{UserID: 99})
				order, err := repo.GetOrder(context.Background(), id)
				require.NoError(t, err)
				require.Equal(t, int64(99), order.UserID)
			},
		},
		{
			name: "GetOrder not found",
			run: func(t *testing.T, repo *InMemoryRepository) {
				_, err := repo.GetOrder(context.Background(), 999)
				require.ErrorIs(t, err, entity.ErrOrderNotFound)
			},
		},
		{
			name: "SetOrderStatus success",
			run: func(t *testing.T, repo *InMemoryRepository) {
				id, _ := repo.CreateOrder(context.Background(), &entity.Order{UserID: 1, Status: entity.OrderStatusAwaitingPay})
				err := repo.SetOrderStatus(context.Background(), id, entity.OrderStatusPayed)
				require.NoError(t, err)
				order, _ := repo.GetOrder(context.Background(), id)
				require.Equal(t, entity.OrderStatusPayed, order.Status)
			},
		},
		{
			name: "SetOrderStatus not found",
			run: func(t *testing.T, repo *InMemoryRepository) {
				err := repo.SetOrderStatus(context.Background(), 999, entity.OrderStatusPayed)
				require.ErrorIs(t, err, entity.ErrOrderNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, NewInMemoryRepository())
		})
	}
}
