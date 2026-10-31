package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type productClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(client productpb.ProductServiceClient) *productClient {
	return &productClient{client: client}
}

func (c *productClient) GetProductInfo(ctx context.Context, sku uint32) (*entity.Product, error) {
	resp, err := c.client.GetProduct(ctx, &productpb.GetProductRequest{Sku: sku})
	if err != nil {
		if st, ok := status.FromError(err); ok { //  ok == true — ошибка пришла от gRPC сервера, можно смотреть на код
			if st.Code() == codes.NotFound {
				return nil, entity.ErrProductNotFound
			}
		}
		return nil, err
	}
	return &entity.Product{
		Name:  resp.GetName(),
		Price: resp.GetPrice(),
	}, nil
}
