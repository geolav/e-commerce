package stocks

type (
	stocksRepository interface {
		// TODO
	}
)

type stocksService struct {
	stocksRepository stocksRepository
}

func NewStocksService(stocksRepository stocksRepository) *stocksService {
	return &stocksService{
		stocksRepository: stocksRepository,
	}
}
