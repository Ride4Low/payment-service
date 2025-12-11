package x402_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ride4Low/payment-service/internal/infrastructure/payment/x402"
)

func TestVerify(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify" {
			t.Errorf("Expected to request '/verify', got: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		resp := x402.VerifyResponse{
			IsValid: true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client with test server URL
	config := &x402.FacilitatorConfig{
		URL: server.URL,
	}
	client := x402.NewFacilitatorClient(config)

	// Test data
	paymentPayload := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload: &x402.ExactEvmPayload{
			Signature: "0xvalidSignature",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        "0xvalidFrom",
				To:          "0xvalidTo",
				Value:       "1000000",
				ValidAfter:  "1745323800",
				ValidBefore: "1745323985",
				Nonce:       "0xvalidNonce",
			},
		},
	}
	paymentRequirements := &x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Resource:          "https://example.com/resource",
		Description:       "Test resource",
		MimeType:          "application/json",
		PayTo:             "0x123",
		MaxTimeoutSeconds: 30,
		Asset:             "0xusdcAddress",
	}

	// Test verify
	resp, err := client.Verify(paymentPayload, paymentRequirements)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !resp.IsValid {
		t.Errorf("Expected valid response, got invalid")
	}
}

func TestSettle(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/settle" {
			t.Errorf("Expected to request '/settle', got: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		resp := x402.SettleResponse{
			Success:     true,
			Transaction: "0xvalidTransaction",
			Network:     "base-sepolia",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client with test server URL
	config := &x402.FacilitatorConfig{
		URL: server.URL,
	}
	client := x402.NewFacilitatorClient(config)

	// Test data
	paymentPayload := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload: &x402.ExactEvmPayload{
			Signature: "0xvalidSignature",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        "0xvalidFrom",
				To:          "0xvalidTo",
				Value:       "1000000",
				ValidAfter:  "1745323800",
				ValidBefore: "1745323985",
				Nonce:       "0xvalidNonce",
			},
		},
	}
	paymentRequirements := &x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Resource:          "https://example.com/resource",
		Description:       "Test resource",
		MimeType:          "application/json",
		PayTo:             "0x123",
		MaxTimeoutSeconds: 30,
		Asset:             "0xusdcAddress",
	}

	// Test settle
	resp, err := client.Settle(paymentPayload, paymentRequirements)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected successful response, got unsuccessful")
	}
	if resp.Transaction != "0xvalidTransaction" {
		t.Errorf("Expected transaction '0xvalidTransaction', got: %s", resp.Transaction)
	}
	if resp.Network != "base-sepolia" {
		t.Errorf("Expected network 'base-sepolia', got: %s", resp.Network)
	}
}

func TestTimeout(t *testing.T) {
	timeoutDuration := time.Millisecond * 100

	// Create test server that takes a while to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * timeoutDuration)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &x402.FacilitatorConfig{
		URL: server.URL,
		Timeout: func() time.Duration {
			return timeoutDuration
		},
	}

	// Create client with test server URL and a timeout option
	client := x402.NewFacilitatorClient(config)

	// Test data
	paymentPayload := &x402.PaymentPayload{}
	paymentRequirements := &x402.PaymentRequirements{}

	// Test verify with timeout
	_, err := client.Verify(paymentPayload, paymentRequirements)
	t.Log(err)
	if err == nil {
		t.Error("Expected timeout error, got err == nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

func TestVerifyWithAuthHeaders(t *testing.T) {
	var capturedAuthHeader string

	// Create test server that captures the auth header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("Authorization")
		resp := x402.VerifyResponse{
			IsValid: true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create mock auth headers function
	createAuthHeaders := func() (map[string]map[string]string, error) {
		return map[string]map[string]string{
			"verify": {"Authorization": "Bearer test-token"},
			"settle": {"Authorization": "Bearer settle-token"},
		}, nil
	}

	// Create client with test server URL and auth headers
	config := &x402.FacilitatorConfig{
		URL:               server.URL,
		CreateAuthHeaders: createAuthHeaders,
	}
	client := x402.NewFacilitatorClient(config)

	// Test verify with auth headers
	paymentPayload := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload: &x402.ExactEvmPayload{
			Signature: "0xvalidSignature",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        "0xvalidFrom",
				To:          "0xvalidTo",
				Value:       "1000000",
				ValidAfter:  "1745323800",
				ValidBefore: "1745323985",
				Nonce:       "0xvalidNonce",
			},
		},
	}
	paymentRequirements := &x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Resource:          "https://example.com/resource",
		Description:       "Test resource",
		MimeType:          "application/json",
		PayTo:             "0x123",
		MaxTimeoutSeconds: 30,
		Asset:             "0xusdcAddress",
	}

	_, err := client.Verify(paymentPayload, paymentRequirements)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that the auth header was set correctly
	expectedAuthHeader := "Bearer test-token"
	if capturedAuthHeader != expectedAuthHeader {
		t.Errorf("Expected auth header '%s', got: '%s'", expectedAuthHeader, capturedAuthHeader)
	}
}

func TestSettleWithAuthHeaders(t *testing.T) {
	var capturedAuthHeader string

	// Create test server that captures the auth header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuthHeader = r.Header.Get("Authorization")
		resp := x402.SettleResponse{
			Success:     true,
			Transaction: "0xvalidTransaction",
			Network:     "base-sepolia",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create mock auth headers function
	createAuthHeaders := func() (map[string]map[string]string, error) {
		return map[string]map[string]string{
			"verify": {"Authorization": "Bearer test-token"},
			"settle": {"Authorization": "Bearer settle-token"},
		}, nil
	}

	// Create client with test server URL and auth headers
	config := &x402.FacilitatorConfig{
		URL:               server.URL,
		CreateAuthHeaders: createAuthHeaders,
	}
	client := x402.NewFacilitatorClient(config)

	// Test settle with auth headers
	paymentPayload := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload: &x402.ExactEvmPayload{
			Signature: "0xvalidSignature",
			Authorization: &x402.ExactEvmPayloadAuthorization{
				From:        "0xvalidFrom",
				To:          "0xvalidTo",
				Value:       "1000000",
				ValidAfter:  "1745323800",
				ValidBefore: "1745323985",
				Nonce:       "0xvalidNonce",
			},
		},
	}
	paymentRequirements := &x402.PaymentRequirements{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Resource:          "https://example.com/resource",
		Description:       "Test resource",
		MimeType:          "application/json",
		PayTo:             "0x123",
		MaxTimeoutSeconds: 30,
		Asset:             "0xusdcAddress",
	}

	_, err := client.Settle(paymentPayload, paymentRequirements)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that the auth header was set correctly
	expectedAuthHeader := "Bearer settle-token"
	if capturedAuthHeader != expectedAuthHeader {
		t.Errorf("Expected auth header '%s', got: '%s'", expectedAuthHeader, capturedAuthHeader)
	}
}
