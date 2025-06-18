package usecase

type ProductTransactionUseCase interface{}
type productTransactionUseCase struct{}

func NewProductTransactionUseCase() ProductTransactionUseCase {
	return &productTransactionUseCase{}
}
