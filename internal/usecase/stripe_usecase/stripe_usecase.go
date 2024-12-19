package stripe_usecase

import (
	"github.com/DanKo-code/Fitness-Center-Abonement/pkg/logger"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
)

type StripeUseCase struct {
	stripeKey string
}

func NewStripeUseCase(stripeKey string) *StripeUseCase {
	return &StripeUseCase{
		stripeKey: stripeKey,
	}
}

func (suc *StripeUseCase) ArchiveStripeProduct(stripePriceId string) error {
	stripe.Key = suc.stripeKey

	priceObject, err := price.Get(stripePriceId, nil)
	if err != nil {
		logger.ErrorLogger.Printf("Error getting stripe id: %v\n", err)
		return err
	}

	productID := priceObject.Product.ID

	params := &stripe.ProductParams{
		Active: stripe.Bool(false), // Устанавливаем активность продукта в false
	}

	updatedProduct, err := product.Update(productID, params)
	if err != nil {
		logger.ErrorLogger.Printf("Ошибка при архивировании продукта: %v\n", err)
		return err
	}

	logger.InfoLogger.Printf("Product %s successfully has been archived\n", updatedProduct.ID)
	return nil
}
