package product

import (
	"context"
	// "errors"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

type InMemoryRepository struct {
	mutex    sync.RWMutex
	products map[uint32]*entity.Product
	stocks   map[uint32]uint64
	nextSKU  uint32
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		products: make(map[uint32]*entity.Product),
		stocks:   make(map[uint32]uint64),
	}
}

// var ErrProductNotFound = errors.New("product not found")

// return only sku because ?? TODO
func (rep *InMemoryRepository) CreateProduct(_ context.Context, name string, price uint32) (uint32, error) {
	rep.mutex.Lock()
	defer rep.mutex.Unlock()
	rep.nextSKU++
	sku := rep.nextSKU
	rep.products[sku] = &entity.Product{SKU: sku, Name: name, Price: price}
	return sku, nil // in future error mb not nil
}

func (rep *InMemoryRepository) GetProduct(_ context.Context, sku uint32) (*entity.Product, error) {
	rep.mutex.RLock()
	defer rep.mutex.RUnlock()
	product, ok := rep.products[sku]
	if !ok {
		return nil, entity.ErrProductNotFound // TODO replace on ErrProductNotFound
	}
	return product, nil
}

func (rep *InMemoryRepository) GetStock(_ context.Context, sku uint32) (uint64, error) {
	rep.mutex.RLock()
	defer rep.mutex.RUnlock()
	return rep.stocks[sku], nil
}

func (rep *InMemoryRepository) SetStock(_ context.Context, sku uint32, count uint64) error {
	rep.mutex.Lock()
	defer rep.mutex.Unlock()
	rep.stocks[sku] = count
	return nil // also in future error mb won't nil
}
