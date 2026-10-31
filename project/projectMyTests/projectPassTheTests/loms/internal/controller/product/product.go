package product

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -source=product.go -destination=mocks/product_mock.go -package=mocks
type (
	productRepository interface {
		CreateProduct(ctx context.Context, name string, price uint32) (uint32, error)
		GetProduct(ctx context.Context, sku uint32) (*entity.Product, error)
	}
)

type productServer struct {
	productRepository productRepository
}

func NewProductServer(productRepository productRepository) *productServer {
	return &productServer{
		productRepository: productRepository,
	}
}

func (s *productServer) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	sku, err := s.productRepository.CreateProduct(ctx, req.GetName(), req.GetPrice())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating product: %v", err)
	}
	return &productpb.CreateProductResponse{Sku: sku}, nil
}

func (s *productServer) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	product, err := s.productRepository.GetProduct(ctx, req.GetSku())
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting product: %v", err)
	}
	return &productpb.GetProductResponse{
		Price: product.Price,
		Name:  product.Name,
	}, nil
}
