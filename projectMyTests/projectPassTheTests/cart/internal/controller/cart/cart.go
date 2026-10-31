package cart

import (
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

var _ cartpb.CartServer = (*cartServer)(nil)

type cartServer struct {
	itemService ItemService
	cartService Service
}

func NewCartServer(
	itemService ItemService,
	cartService Service,
) *cartServer {
	return &cartServer{
		itemService: itemService,
		cartService: cartService,
	}
}
