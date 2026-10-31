package controller

import (
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
)

type API struct {
	Cart cartpb.CartServer
}

func New(cart cartpb.CartServer) *API {
	return &API{Cart: cart}
}
