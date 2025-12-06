package stripe

import (
	"context"
	"errors"
	"testing"

	"github.com/stripe/stripe-go/v81"
)

func TestProvider_CreatePaymentSession_Success(t *testing.T) {
	config := PaymentConfig{
		StripeSecretKey: "sk_test_123",
		SuccessURL:      "http://localhost:3000/success",
		CancelURL:       "http://localhost:3000/cancel",
	}

	expectedSessionID := "cs_test_session_123"

	mockCreator := func(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
		// Verify params
		if *params.SuccessURL != config.SuccessURL {
			t.Errorf("expected success URL %s, got %s", config.SuccessURL, *params.SuccessURL)
		}
		if *params.CancelURL != config.CancelURL {
			t.Errorf("expected cancel URL %s, got %s", config.CancelURL, *params.CancelURL)
		}

		// Verify amount (unit amount is in cents)
		expectedAmount := int64(1000)
		if *params.LineItems[0].PriceData.UnitAmount != expectedAmount {
			t.Errorf("expected amount %d, got %d", expectedAmount, *params.LineItems[0].PriceData.UnitAmount)
		}

		return &stripe.CheckoutSession{ID: expectedSessionID}, nil
	}

	provider := NewProviderWithCreator(config, mockCreator)

	ctx := context.Background()
	sessionID, err := provider.CreatePaymentSession(ctx, 1000, "usd", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sessionID != expectedSessionID {
		t.Errorf("expected session ID %s, got %s", expectedSessionID, sessionID)
	}
}

func TestProvider_CreatePaymentSession_LinkItems(t *testing.T) {
	config := PaymentConfig{
		StripeSecretKey: "sk_test_123",
	}

	mockCreator := func(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
		if len(params.LineItems) != 1 {
			t.Errorf("expected 1 line item, got %d", len(params.LineItems))
		}

		item := params.LineItems[0]
		if *item.Quantity != 1 {
			t.Errorf("expected quantity 1, got %d", *item.Quantity)
		}

		if *item.PriceData.Currency != "eur" {
			t.Errorf("expected currency eur, got %s", *item.PriceData.Currency)
		}

		return &stripe.CheckoutSession{ID: "sess_123"}, nil
	}

	provider := NewProviderWithCreator(config, mockCreator)
	provider.CreatePaymentSession(context.Background(), 500, "eur", nil)
}

func TestProvider_CreatePaymentSession_Failure(t *testing.T) {
	config := PaymentConfig{
		StripeSecretKey: "sk_test_123",
	}

	expectedErr := errors.New("stripe api down")

	mockCreator := func(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
		return nil, expectedErr
	}

	provider := NewProviderWithCreator(config, mockCreator)

	ctx := context.Background()
	_, err := provider.CreatePaymentSession(ctx, 1000, "usd", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
