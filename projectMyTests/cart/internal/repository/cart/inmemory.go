package cart

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

type inMemoryCartRepository struct {
	mutex sync.RWMutex
	items map[int64][]entity.OrderItem // не по указателю потому что это маленькая структура TODO или может лучше по указателю?
}

func NewInMemoryRepository() *inMemoryCartRepository {
	return &inMemoryCartRepository{
		items: make(map[int64][]entity.OrderItem),
	}
}

func (repo *inMemoryCartRepository) AddItem(_ context.Context, userID int64, orderItem entity.OrderItem) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for i, item := range repo.items[userID] {
		if item.SKU == orderItem.SKU {
			repo.items[userID][i].Count += orderItem.Count
			return nil
		}
	}
	repo.items[userID] = append(repo.items[userID], orderItem)
	return nil
}

func (repo *inMemoryCartRepository) DeleteItem(_ context.Context, userID int64, sku uint32) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for i, item := range repo.items[userID] {
		if item.SKU == sku {
			repo.items[userID] = append(repo.items[userID][:i], repo.items[userID][i+1:]...)
			return nil
		}
	}
	return nil
}

func (repo *inMemoryCartRepository) ListItems(_ context.Context, userID int64) ([]entity.OrderItem, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()
	return repo.items[userID], nil // мб потом будем ошибку возвращать
}

func (repo *inMemoryCartRepository) ClearCart(_ context.Context, userID int64) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	delete(repo.items, userID)
	return nil // мб потом будем ошибку возвращать
}
