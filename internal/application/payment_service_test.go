package application

import (
	"context"
	"errors"
	"testing"
)

// mockPaymentProvider is a mock implementation of PaymentProvider for testing
type mockPaymentProvider struct {
	sessionID string
	err       error
}

func (m *mockPaymentProvider) CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.sessionID, nil
}

func TestNewPaymentService(t *testing.T) {
	provider := &mockPaymentProvider{}
	svc := NewPaymentService(provider)

	if svc == nil {
		t.Fatal("expected non-nil PaymentService")
	}
}

func TestPaymentService_CreatePaymentSession_Success(t *testing.T) {
	provider := &mockPaymentProvider{sessionID: "cs_test_session_123"}
	svc := NewPaymentService(provider)

	ctx := context.Background()
	intent, err := svc.CreatePaymentSession(ctx, "trip-1", "user-1", "driver-1", 1000, "usd")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if intent == nil {
		t.Fatal("expected non-nil PaymentIntent")
	}

	if intent.TripID != "trip-1" {
		t.Errorf("expected TripID 'trip-1', got '%s'", intent.TripID)
	}

	if intent.UserID != "user-1" {
		t.Errorf("expected UserID 'user-1', got '%s'", intent.UserID)
	}

	if intent.DriverID != "driver-1" {
		t.Errorf("expected DriverID 'driver-1', got '%s'", intent.DriverID)
	}

	if intent.Amount != 1000 {
		t.Errorf("expected Amount 1000, got %d", intent.Amount)
	}

	if intent.Currency != "usd" {
		t.Errorf("expected Currency 'usd', got '%s'", intent.Currency)
	}

	if intent.StripeSessionID != "cs_test_session_123" {
		t.Errorf("expected StripeSessionID 'cs_test_session_123', got '%s'", intent.StripeSessionID)
	}

	if intent.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestPaymentService_CreatePaymentSession_ProviderError(t *testing.T) {
	providerErr := errors.New("stripe api error")
	provider := &mockPaymentProvider{err: providerErr}
	svc := NewPaymentService(provider)

	ctx := context.Background()
	intent, err := svc.CreatePaymentSession(ctx, "trip-1", "user-1", "driver-1", 1000, "usd")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if intent != nil {
		t.Error("expected nil PaymentIntent on error")
	}

	if !errors.Is(err, providerErr) {
		t.Errorf("expected error to be '%v', got '%v'", providerErr, err)
	}
}
