package order

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

func TestCreateOrder(t *testing.T) {
	repo := NewInMemoryRepository()
	order := &entity.Order{UserID: 1, Status: entity.OrderStatusAwaitingPay}

	id, err := repo.CreateOrder(context.Background(), order)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)
}

func TestGetOrder_Success(t *testing.T) {
	repo := NewInMemoryRepository()
	order := &entity.Order{UserID: 1, Status: entity.OrderStatusAwaitingPay}
	id, _ := repo.CreateOrder(context.Background(), order)

	got, err := repo.GetOrder(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.UserID)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	_, err := repo.GetOrder(context.Background(), 999)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}

func TestSetOrderStatus(t *testing.T) {
	repo := NewInMemoryRepository()
	order := &entity.Order{UserID: 1, Status: entity.OrderStatusAwaitingPay}
	id, _ := repo.CreateOrder(context.Background(), order)

	err := repo.SetOrderStatus(context.Background(), id, entity.OrderStatusPayed)
	require.NoError(t, err)

	got, _ := repo.GetOrder(context.Background(), id)
	require.Equal(t, entity.OrderStatusPayed, got.Status)
}

func TestSetOrderStatus_NotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SetOrderStatus(context.Background(), 999, entity.OrderStatusPayed)
	require.ErrorIs(t, err, entity.ErrOrderNotFound)
}
