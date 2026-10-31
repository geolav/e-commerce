package controller

import (
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

type API struct {
	Loms    pbloms.LomsServer // без указателей, потому что интерфейсы итак по ссылке хранятся
	Product productpb.ProductServiceServer
	Stocks  stockspb.StocksServer
}

func New(loms pbloms.LomsServer,
	product productpb.ProductServiceServer,
	stocks stockspb.StocksServer) *API {
	return &API{
		Loms:    loms,
		Product: product,
		Stocks:  stocks,
	}
}
