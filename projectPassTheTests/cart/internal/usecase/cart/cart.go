package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mock.go -package=mocks

type (
	cartRepository interface {
		ListItems(ctx context.Context, userID int64) ([]entity.OrderItem, error)
		ClearCart(ctx context.Context, userID int64) error
	}

	productClient interface {
		GetProductInfo(ctx context.Context, sku uint32) (*entity.Product, error)
	}

	lomsClient interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error)
	}
)

type cartService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewCartService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *cartService {
	return &cartService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}

func (s *cartService) ListCart(ctx context.Context, userID int64) ([]entity.OrderItem, uint32, error) {
	items, err := s.cartRepository.ListItems(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	var totalPrice uint32
	for i, item := range items {
		product, err := s.productClient.GetProductInfo(ctx, item.SKU)
		if err != nil {
			return nil, 0, err
		}
		items[i].Name = product.Name
		items[i].Price = product.Price
		totalPrice += product.Price * item.Count
	}
	return items, totalPrice, nil
}

func (s *cartService) ClearCart(ctx context.Context, userID int64) error {
	return s.cartRepository.ClearCart(ctx, userID)
}

func (s *cartService) CheckoutCart(ctx context.Context, userID int64) (int64, error) {
	items, _, err := s.ListCart(ctx, userID)
	if err != nil {
		return 0, err
	}

	orderID, err := s.lomsClient.CreateOrder(ctx, userID, items)
	if err != nil {
		return 0, err
	}

	if err := s.cartRepository.ClearCart(ctx, userID); err != nil {
		return 0, err
	}

	return orderID, nil
}
