package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	pbloms "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	stocksbp "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

type LomsClient struct {
	lomsClient   pbloms.LomsClient
	stocksClient stocksbp.StocksClient
}

func NewLOMSClient(lomsClient pbloms.LomsClient, stocksClient stocksbp.StocksClient) *LomsClient {
	return &LomsClient{
		lomsClient:   lomsClient,
		stocksClient: stocksClient,
	}
}

// CreateOrder отправляет заказ в LOMS через gRPC.
//
// Важная деталь реализации:
// pbItems создаётся через make([]*pbloms.Item, 0, len(items)) — с длиной 0 и capacity = len(items).
// Это важно, потому что:
//  1. make([]*T, len) создаёт nil-элементы в количестве len штук, и append добавляет новые после них.
//     Тогда, если использовать именно такой вариант, LOMS приходил бы слайс вида [nil, nil, ..., реальный_элемент].
//  2. nil-элементы при маршалинге в protobuf дают пустые Item с SKU=0, Count=0,
//     что привело бы (и в процессе тестирования приводило) к ошибкам резервирования (попытка зарезервировать 0 единиц SKU 0).
//  3. С длиной 0 слайс содержит только реальные элементы без мусора.
func (c *LomsClient) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	pbItems := make([]*pbloms.Item, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &pbloms.Item{
			Sku:   item.SKU,
			Count: item.Count,
		})
	}
	resp, err := c.lomsClient.CreateOrder(ctx, &pbloms.CreateOrderRequest{
		UserId: userID,
		Items:  pbItems,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetOrderId(), nil
}

func (c *LomsClient) GetStocks(ctx context.Context, sku uint32) (uint64, error) {
	resp, err := c.stocksClient.GetStock(ctx, &stocksbp.GetStockRequest{
		Sku: sku,
	})
	if err != nil {
		return 0, err
	}
	return resp.GetCount(), nil
}
