package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type PaymentConfig struct {
	StripeSecretKey     string `json:"stripeSecretKey"`
	StripeWebhookSecret string `json:"stripeWebhookSecret"`
	Currency            string `json:"currency"`
	SuccessURL          string `json:"successURL"`
	CancelURL           string `json:"cancelURL"`
}

// SessionCreator defines a function that creates a checkout session
type SessionCreator func(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error)

// Provider implements application.PaymentProvider for Stripe
type Provider struct {
	config        PaymentConfig
	createSession SessionCreator
}

// NewProvider creates a new Stripe payment provider
func NewProvider(config PaymentConfig) *Provider {
	stripe.Key = config.StripeSecretKey
	return &Provider{
		config:        config,
		createSession: session.New,
	}
}

// NewProviderWithCreator creates a new Stripe payment provider with a custom session creator (for testing)
func NewProviderWithCreator(config PaymentConfig, creator SessionCreator) *Provider {
	stripe.Key = config.StripeSecretKey
	return &Provider{
		config:        config,
		createSession: creator,
	}
}

// CreatePaymentSession creates a Stripe checkout session
func (p *Provider) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(p.config.SuccessURL),
		CancelURL:  stripe.String(p.config.CancelURL),
		Metadata:   metadata,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Test Product"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	result, err := p.createSession(params)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}
