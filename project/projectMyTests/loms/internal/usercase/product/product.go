package product

type (
	productRepository interface {
		// TODO
	}
)

type productService struct {
	productRepository productRepository
}

func NewProductService(productRepository productRepository) *productService {
	return &productService{
		productRepository: productRepository,
	}
}
