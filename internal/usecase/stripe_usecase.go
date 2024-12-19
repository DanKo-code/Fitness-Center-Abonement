package usecase

type StripeUseCase interface {
	ArchiveStripeProduct(stripePriceId string) error
}
