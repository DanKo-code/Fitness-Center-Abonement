package usecase

type StripeUseCase interface {
	ArchiveStripeProduct(stripePriceId string) error
	CreateStripeProductAndPrice(name string, amount int64, currency string) (string, error)
}
