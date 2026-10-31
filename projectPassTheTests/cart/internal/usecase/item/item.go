package item

import (
	"context"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

//go:generate mockgen -source=item.go -destination=mocks/item_mock.go -package=mocks

type (
	cartRepository interface {
		AddItem(ctx context.Context, userID int64, item entity.OrderItem) error
		DeleteItem(ctx context.Context, userID int64, sku uint32) error
	}

	productClient interface {
		GetProductInfo(ctx context.Context, sku uint32) (*entity.Product, error)
	}

	lomsClient interface {
		GetStocks(ctx context.Context, sku uint32) (uint64, error)
	}
)

type itemService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewItemService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *itemService {
	return &itemService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}

func (s *itemService) AddItem(ctx context.Context, userID int64, sku uint32, count uint32) error {
	_, err := s.productClient.GetProductInfo(ctx, sku)
	if err != nil {
		return fmt.Errorf("could not get product info: %w", entity.ErrProductNotFound) // TODO mb: entity.ErrProductNotFound
	}
	lefts, err := s.lomsClient.GetStocks(ctx, sku)
	if err != nil {
		return fmt.Errorf("could not get stocks: %w", err)
	}
	if uint64(count) > lefts {
		return fmt.Errorf("insufficient stock: %w", entity.ErrInsufficientStock) // TODO mb entity.ErrInsufficientStock
	}
	return s.cartRepository.AddItem(ctx, userID, entity.OrderItem{
		SKU:   sku,
		Count: count,
	})
}

func (s *itemService) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	return s.cartRepository.DeleteItem(ctx, userID, sku)
}
