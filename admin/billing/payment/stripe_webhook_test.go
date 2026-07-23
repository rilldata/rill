package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/jobs"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
)

const testStripeWebhookSecret = "whsec_bill_002"

type paymentMethodAddedCall struct {
	methodID          string
	paymentCustomerID string
	typ               string
	eventTime         time.Time
}

type paymentMethodRemovedCall struct {
	methodID          string
	paymentCustomerID string
	eventTime         time.Time
}

type customerAddressUpdatedCall struct {
	paymentCustomerID string
	eventTime         time.Time
}

type stripeWebhookJobs struct {
	jobs.Client

	paymentMethodAddedCalls     []paymentMethodAddedCall
	paymentMethodRemovedCalls   []paymentMethodRemovedCall
	customerAddressUpdatedCalls []customerAddressUpdatedCall
	paymentMethodAddedErr       error
	paymentMethodRemovedErr     error
	customerAddressUpdatedErr   error
}

func (j *stripeWebhookJobs) PaymentMethodAdded(_ context.Context, methodID, paymentCustomerID, typ string, eventTime time.Time) (*jobs.InsertResult, error) {
	j.paymentMethodAddedCalls = append(j.paymentMethodAddedCalls, paymentMethodAddedCall{
		methodID:          methodID,
		paymentCustomerID: paymentCustomerID,
		typ:               typ,
		eventTime:         eventTime,
	})
	return &jobs.InsertResult{}, j.paymentMethodAddedErr
}

func (j *stripeWebhookJobs) PaymentMethodRemoved(_ context.Context, methodID, paymentCustomerID string, eventTime time.Time) (*jobs.InsertResult, error) {
	j.paymentMethodRemovedCalls = append(j.paymentMethodRemovedCalls, paymentMethodRemovedCall{
		methodID:          methodID,
		paymentCustomerID: paymentCustomerID,
		eventTime:         eventTime,
	})
	return &jobs.InsertResult{}, j.paymentMethodRemovedErr
}

func (j *stripeWebhookJobs) CustomerAddressUpdated(_ context.Context, paymentCustomerID string, eventTime time.Time) (*jobs.InsertResult, error) {
	j.customerAddressUpdatedCalls = append(j.customerAddressUpdatedCalls, customerAddressUpdatedCall{
		paymentCustomerID: paymentCustomerID,
		eventTime:         eventTime,
	})
	return &jobs.InsertResult{}, j.customerAddressUpdatedErr
}

func (j *stripeWebhookJobs) callCount() int {
	return len(j.paymentMethodAddedCalls) + len(j.paymentMethodRemovedCalls) + len(j.customerAddressUpdatedCalls)
}

func TestStripeWebhookSignedEventsEnqueueParsedJobs(t *testing.T) {
	// Stripe's signing helper keeps the tests on the same verification boundary as production requests.
	tests := []struct {
		name       string
		payload    []byte
		assertCall func(*testing.T, *stripeWebhookJobs)
	}{
		{
			name: "payment method attached",
			payload: stripeEventPayload(t, map[string]any{
				"id":      "evt_attached",
				"type":    "payment_method.attached",
				"created": int64(1_717_171_717),
				"data": map[string]any{"object": map[string]any{
					"id":       "pm_attached",
					"object":   "payment_method",
					"created":  int64(1_700_000_111),
					"customer": "cus_attached",
					"type":     "card",
				}},
			}),
			assertCall: func(t *testing.T, got *stripeWebhookJobs) {
				require.Equal(t, []paymentMethodAddedCall{{
					methodID:          "pm_attached",
					paymentCustomerID: "cus_attached",
					typ:               "card",
					eventTime:         time.Unix(1_700_000_111, 0),
				}}, got.paymentMethodAddedCalls)
			},
		},
		{
			name: "payment method detached",
			payload: stripeEventPayload(t, map[string]any{
				"id":      "evt_detached",
				"type":    "payment_method.detached",
				"created": int64(1_717_171_718),
				"data": map[string]any{
					"object": map[string]any{
						"id":      "pm_detached",
						"object":  "payment_method",
						"created": int64(1_700_000_222),
						"type":    "us_bank_account",
					},
					"previous_attributes": map[string]any{"customer": "cus_detached"},
				},
			}),
			assertCall: func(t *testing.T, got *stripeWebhookJobs) {
				require.Equal(t, []paymentMethodRemovedCall{{
					methodID:          "pm_detached",
					paymentCustomerID: "cus_detached",
					eventTime:         time.Unix(1_700_000_222, 0),
				}}, got.paymentMethodRemovedCalls)
			},
		},
		{
			name: "customer address updated",
			payload: stripeEventPayload(t, map[string]any{
				"id":      "evt_customer",
				"type":    "customer.updated",
				"created": int64(1_700_000_333),
				"data": map[string]any{
					"object": map[string]any{
						"id":     "cus_updated",
						"object": "customer",
						"address": map[string]any{
							"country": "US",
							"line1":   "123 Main Street",
						},
					},
					"previous_attributes": map[string]any{"address": nil},
				},
			}),
			assertCall: func(t *testing.T, got *stripeWebhookJobs) {
				require.Equal(t, []customerAddressUpdatedCall{{
					paymentCustomerID: "cus_updated",
					eventTime:         time.Unix(1_700_000_333, 0),
				}}, got.customerAddressUpdatedCalls)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Every recognized, authenticated event must cross into the jobs system exactly once with its parsed identifiers and timestamp.
			jobClient := &stripeWebhookJobs{}
			response := serveStripeWebhook(t, jobClient, tt.payload, signedStripeHeader(tt.payload, testStripeWebhookSecret))

			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, 1, jobClient.callCount())
			tt.assertCall(t, jobClient)
		})
	}
}

func TestStripeWebhookDetachedPreviousCustomerShapes(t *testing.T) {
	// Detached payment methods omit their current customer, making the signed previous value the only usable association.
	tests := []struct {
		name             string
		previousCustomer any
		wantCustomerID   string
		wantCalls        int
	}{
		{name: "string", previousCustomer: "cus_previous", wantCustomerID: "cus_previous", wantCalls: 1},
		{name: "null", previousCustomer: nil},
		{name: "object", previousCustomer: map[string]any{"id": "cus_object"}},
		{name: "number", previousCustomer: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Webhook metadata is untrusted input even after authentication; schema drift must not panic or invent a customer ID.
			payload := stripeEventPayload(t, map[string]any{
				"id":      "evt_previous_customer_" + tt.name,
				"type":    "payment_method.detached",
				"created": int64(1_700_000_400),
				"data": map[string]any{
					"object": map[string]any{
						"id":      "pm_previous_customer",
						"object":  "payment_method",
						"created": int64(1_700_000_401),
					},
					"previous_attributes": map[string]any{"customer": tt.previousCustomer},
				},
			})
			jobClient := &stripeWebhookJobs{}
			response := serveStripeWebhook(t, jobClient, payload, signedStripeHeader(payload, testStripeWebhookSecret))

			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, tt.wantCalls, jobClient.callCount())
			if tt.wantCalls == 1 {
				require.Equal(t, tt.wantCustomerID, jobClient.paymentMethodRemovedCalls[0].paymentCustomerID)
			}
		})
	}
}

func TestStripeWebhookRejectsUntrustedOrInvalidBodies(t *testing.T) {
	// Rejected requests must stop at the HTTP boundary and never reach any asynchronous billing work.
	t.Run("bad signature", func(t *testing.T) {
		// A syntactically valid event signed with another secret must be indistinguishable from unauthenticated input.
		payload := stripeEventPayload(t, map[string]any{
			"id": "evt_bad_signature", "type": "payment_method.attached", "created": int64(1_700_000_500),
			"data": map[string]any{"object": map[string]any{
				"id": "pm_bad_signature", "object": "payment_method", "customer": "cus_bad_signature", "type": "card",
			}},
		})
		jobClient := &stripeWebhookJobs{}
		response := serveStripeWebhook(t, jobClient, payload, signedStripeHeader(payload, "whsec_wrong"))

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, jobClient.callCount())
	})

	t.Run("signed malformed JSON", func(t *testing.T) {
		// Signing malformed bytes proves authentication alone cannot bypass event decoding.
		payload := []byte(`{"id":"evt_malformed"`)
		jobClient := &stripeWebhookJobs{}
		response := serveStripeWebhook(t, jobClient, payload, signedStripeHeader(payload, testStripeWebhookSecret))

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, jobClient.callCount())
	})

	t.Run("body exceeds limit", func(t *testing.T) {
		// One byte beyond the explicit cap verifies oversized bodies fail before signature parsing or enqueueing.
		payload := bytes.Repeat([]byte("x"), int(maxBodyBytes)+1)
		jobClient := &stripeWebhookJobs{}
		response := serveStripeWebhook(t, jobClient, payload, signedStripeHeader(payload, testStripeWebhookSecret))

		require.Equal(t, http.StatusServiceUnavailable, response.Code)
		require.Zero(t, jobClient.callCount())
	})
}

func TestStripeWebhookReturnsServerErrorWhenEnqueueFails(t *testing.T) {
	// A failed handoff must not be acknowledged, allowing Stripe to retry without a duplicate enqueue in this request.
	tests := []struct {
		name      string
		payload   []byte
		configure func(*stripeWebhookJobs)
	}{
		{
			name: "payment method attached",
			payload: stripeEventPayload(t, map[string]any{
				"id": "evt_failed_attached", "type": "payment_method.attached", "created": int64(1_700_000_600),
				"data": map[string]any{"object": map[string]any{
					"id": "pm_failed_attached", "object": "payment_method", "created": int64(1_700_000_601), "customer": "cus_failed_attached", "type": "card",
				}},
			}),
			configure: func(j *stripeWebhookJobs) { j.paymentMethodAddedErr = errors.New("queue unavailable") },
		},
		{
			name: "payment method detached",
			payload: stripeEventPayload(t, map[string]any{
				"id": "evt_failed_detached", "type": "payment_method.detached", "created": int64(1_700_000_610),
				"data": map[string]any{
					"object":              map[string]any{"id": "pm_failed_detached", "object": "payment_method", "created": int64(1_700_000_611)},
					"previous_attributes": map[string]any{"customer": "cus_failed_detached"},
				},
			}),
			configure: func(j *stripeWebhookJobs) { j.paymentMethodRemovedErr = errors.New("queue unavailable") },
		},
		{
			name: "customer address updated",
			payload: stripeEventPayload(t, map[string]any{
				"id": "evt_failed_customer", "type": "customer.updated", "created": int64(1_700_000_620),
				"data": map[string]any{
					"object":              map[string]any{"id": "cus_failed", "object": "customer", "address": map[string]any{"country": "US"}},
					"previous_attributes": map[string]any{"address": nil},
				},
			}),
			configure: func(j *stripeWebhookJobs) { j.customerAddressUpdatedErr = errors.New("queue unavailable") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each job path gets one handoff attempt; returning 500 preserves Stripe's retry signal when that attempt fails.
			jobClient := &stripeWebhookJobs{}
			tt.configure(jobClient)
			response := serveStripeWebhook(t, jobClient, tt.payload, signedStripeHeader(tt.payload, testStripeWebhookSecret))

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Equal(t, 1, jobClient.callCount())
		})
	}
}

func stripeEventPayload(t *testing.T, event map[string]any) []byte {
	t.Helper()
	event["object"] = "event"
	payload, err := json.Marshal(event)
	require.NoError(t, err)
	return payload
}

func signedStripeHeader(payload []byte, secret string) string {
	return webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  secret,
	}).Header
}

func serveStripeWebhook(t *testing.T, jobClient jobs.Client, payload []byte, signature string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewStripe(zap.NewNop(), "", testStripeWebhookSecret).WebhookHandlerFunc(context.Background(), jobClient)
	require.NotNil(t, handler)

	request := httptest.NewRequest(http.MethodPost, "/stripe/webhook", bytes.NewReader(payload))
	request.Header.Set("Stripe-Signature", signature)
	response := httptest.NewRecorder()
	require.NotPanics(t, func() {
		handler.ServeHTTP(response, request)
	})
	return response
}
