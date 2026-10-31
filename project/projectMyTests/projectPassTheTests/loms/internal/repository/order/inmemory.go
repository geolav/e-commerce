package order

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

type InMemoryRepository struct {
	mutex  sync.RWMutex
	orders map[int64]*entity.Order
	nextID int64
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		orders: make(map[int64]*entity.Order),
	}
}

func (rep *InMemoryRepository) CreateOrder(_ context.Context, order *entity.Order) (int64, error) {
	rep.mutex.Lock()
	defer rep.mutex.Unlock()
	rep.nextID++
	order.ID = rep.nextID
	rep.orders[order.ID] = order
	return rep.nextID, nil
}

func (rep *InMemoryRepository) GetOrder(_ context.Context, orderID int64) (*entity.Order, error) {
	rep.mutex.RLock()
	defer rep.mutex.RUnlock()
	order, ok := rep.orders[orderID]
	if !ok {
		return nil, entity.ErrOrderNotFound
	}
	return order, nil
}

func (rep *InMemoryRepository) SetOrderStatus(_ context.Context, orderID int64, status entity.OrderStatus) error {
	rep.mutex.Lock()
	defer rep.mutex.Unlock()
	order, ok := rep.orders[orderID]
	if !ok {
		return entity.ErrOrderNotFound
	}
	order.Status = status
	return nil
}
