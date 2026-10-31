package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/cart_mock.go -package=mocks

type (
	ItemService interface {
		AddItem(ctx context.Context, userID int64, sku, count uint32) error
		DeleteItem(ctx context.Context, userID int64, sku uint32) error
	}

	Service interface {
		ListCart(ctx context.Context, userID int64) ([]entity.OrderItem, uint32, error)
		ClearCart(ctx context.Context, userID int64) error
		CheckoutCart(ctx context.Context, userID int64) (int64, error)
	}
)
