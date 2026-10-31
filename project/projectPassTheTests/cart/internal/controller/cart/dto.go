package cart

import (
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

func toProtoItem(item entity.OrderItem) *cartpb.Item {
	return &cartpb.Item{
		Sku:   item.SKU,
		Count: item.Count,
		Name:  item.Name,
		Price: item.Price,
	}
}
