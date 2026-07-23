package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/jobs"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testOrbWebhookSecret = "orb_webhook_secret_bill_001"

var (
	testOrbDueDate       = time.Date(2026, time.July, 29, 12, 30, 0, 0, time.UTC)
	testOrbPaymentFailed = time.Date(2026, time.July, 22, 8, 15, 30, 123000000, time.UTC)
)

type orbPaymentSuccessCall struct {
	billingCustomerID string
	invoiceID         string
}

type orbPaymentFailedCall struct {
	billingCustomerID string
	invoiceID         string
	invoiceNumber     string
	invoiceURL        string
	amount            string
	currency          string
	dueDate           time.Time
	failedAt          time.Time
}

// orbWebhookJobs deliberately implements only the queue calls reachable from
// this webhook. The embedded interface makes any unexpected queue dependency
// fail loudly instead of turning the test fake into a second jobs.Client.
type orbWebhookJobs struct {
	jobs.Client

	paymentSuccessCalls        []orbPaymentSuccessCall
	paymentFailedCalls         []orbPaymentFailedCall
	planChangedCalls           []string
	creditBalanceDroppedCalls  []string
	creditBalanceDepletedCalls []string
	queueErr                   error
	duplicateOnRepeat          bool
}

func (j *orbWebhookJobs) PaymentSuccess(_ context.Context, billingCustomerID, invoiceID string) (*jobs.InsertResult, error) {
	j.paymentSuccessCalls = append(j.paymentSuccessCalls, orbPaymentSuccessCall{
		billingCustomerID: billingCustomerID,
		invoiceID:         invoiceID,
	})
	return j.insertResult()
}

func (j *orbWebhookJobs) PaymentFailed(_ context.Context, billingCustomerID, invoiceID, invoiceNumber, invoiceURL, amount, currency string, dueDate, failedAt time.Time) (*jobs.InsertResult, error) {
	j.paymentFailedCalls = append(j.paymentFailedCalls, orbPaymentFailedCall{
		billingCustomerID: billingCustomerID,
		invoiceID:         invoiceID,
		invoiceNumber:     invoiceNumber,
		invoiceURL:        invoiceURL,
		amount:            amount,
		currency:          currency,
		dueDate:           dueDate,
		failedAt:          failedAt,
	})
	return j.insertResult()
}

func (j *orbWebhookJobs) PlanChanged(_ context.Context, billingCustomerID string) (*jobs.InsertResult, error) {
	j.planChangedCalls = append(j.planChangedCalls, billingCustomerID)
	return j.insertResult()
}

func (j *orbWebhookJobs) CreditBalanceDropped(_ context.Context, billingCustomerID string) (*jobs.InsertResult, error) {
	j.creditBalanceDroppedCalls = append(j.creditBalanceDroppedCalls, billingCustomerID)
	return j.insertResult()
}

func (j *orbWebhookJobs) CreditBalanceDepleted(_ context.Context, billingCustomerID string) (*jobs.InsertResult, error) {
	j.creditBalanceDepletedCalls = append(j.creditBalanceDepletedCalls, billingCustomerID)
	return j.insertResult()
}

func (j *orbWebhookJobs) insertResult() (*jobs.InsertResult, error) {
	if j.queueErr != nil {
		return nil, j.queueErr
	}
	return &jobs.InsertResult{Duplicate: j.duplicateOnRepeat && j.callCount() > 1}, nil
}

func (j *orbWebhookJobs) callCount() int {
	return len(j.paymentSuccessCalls) +
		len(j.paymentFailedCalls) +
		len(j.planChangedCalls) +
		len(j.creditBalanceDroppedCalls) +
		len(j.creditBalanceDepletedCalls)
}

func TestOrbWebhookSignedInterestingEvents(t *testing.T) {
	// Map every supported signed event to its durable job payload and characterize intentionally ignored events.
	tests := []struct {
		name       string
		eventType  string
		payload    func(*testing.T, string) []byte
		assertJobs func(*testing.T, *orbWebhookJobs)
	}{
		{
			name:      "invoice payment succeeded",
			eventType: "invoice.payment_succeeded",
			payload:   orbInvoicePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []orbPaymentSuccessCall{{
					billingCustomerID: "org_invoice_customer",
					invoiceID:         "inv_bill_001",
				}}, got.paymentSuccessCalls)
			},
		},
		{
			name:      "invoice payment failed",
			eventType: "invoice.payment_failed",
			payload:   orbInvoicePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []orbPaymentFailedCall{{
					billingCustomerID: "org_invoice_customer",
					invoiceID:         "inv_bill_001",
					invoiceNumber:     "RILL-000042",
					invoiceURL:        "https://billing.example.test/invoices/inv_bill_001",
					amount:            "42.75",
					currency:          "USD",
					dueDate:           testOrbDueDate,
					failedAt:          testOrbPaymentFailed,
				}}, got.paymentFailedCalls)
			},
		},
		{
			name:      "invoice issue failed",
			eventType: "invoice.issue_failed",
			payload:   orbInvoicePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Zero(t, got.callCount())
			},
		},
		{
			name:      "subscription started",
			eventType: "subscription.started",
			payload:   orbSubscriptionPayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_subscription_customer"}, got.planChangedCalls)
			},
		},
		{
			name:      "subscription ended",
			eventType: "subscription.ended",
			payload:   orbSubscriptionPayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_subscription_customer"}, got.planChangedCalls)
			},
		},
		{
			name:      "subscription plan changed",
			eventType: "subscription.plan_changed",
			payload:   orbSubscriptionPayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_subscription_customer"}, got.planChangedCalls)
			},
		},
		{
			name:      "credit balance dropped",
			eventType: "customer.credit_balance_dropped",
			payload:   orbCreditBalancePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_credit_customer"}, got.creditBalanceDroppedCalls)
			},
		},
		{
			name:      "credit balance depleted",
			eventType: "customer.credit_balance_depleted",
			payload:   orbCreditBalancePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_credit_customer"}, got.creditBalanceDepletedCalls)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Every state-changing path is exercised with an HMAC over the exact
			// transmitted JSON, not a stubbed verifier or a made-up digest.
			payload := tt.payload(t, tt.eventType)
			timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
			jobClient := &orbWebhookJobs{}
			response := serveOrbWebhook(t, jobClient, payload, timestamp, orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

			require.Equal(t, http.StatusOK, response.Code)
			tt.assertJobs(t, jobClient)
			if tt.eventType != "invoice.issue_failed" {
				require.Equal(t, 1, jobClient.callCount())
			}
		})
	}
}

func TestOrbWebhookSignatureHeaders(t *testing.T) {
	// Signature parsing must reject incomplete candidates while accepting one valid key-rotation candidate.
	payload := orbInvoicePayload(t, "invoice.payment_succeeded")
	timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
	valid := orbSignedHeader(payload, timestamp, testOrbWebhookSecret)
	wrong := orbSignedHeader(payload, timestamp, "wrong-secret")

	tests := []struct {
		name       string
		timestamp  string
		signatures []string
		wantStatus int
		wantCalls  int
	}{
		{name: "missing signature", timestamp: timestamp, wantStatus: http.StatusBadRequest},
		{name: "missing timestamp", signatures: []string{valid}, wantStatus: http.StatusBadRequest},
		{name: "invalid timestamp", timestamp: "not-a-timestamp", signatures: []string{valid}, wantStatus: http.StatusBadRequest},
		{name: "invalid hex", timestamp: timestamp, signatures: []string{"v1=not-hex"}, wantStatus: http.StatusBadRequest},
		{name: "wrong digest", timestamp: timestamp, signatures: []string{wrong}, wantStatus: http.StatusBadRequest},
		{name: "unsupported version", timestamp: timestamp, signatures: []string{"v2=" + valid[3:]}, wantStatus: http.StatusBadRequest},
		{name: "all repeated signatures invalid", timestamp: timestamp, signatures: []string{wrong, "v1=00"}, wantStatus: http.StatusBadRequest},
		{name: "valid repeated signature", timestamp: timestamp, signatures: []string{wrong, valid}, wantStatus: http.StatusOK, wantCalls: 1},
		{name: "valid comma separated signature", timestamp: timestamp, signatures: []string{wrong + ", " + valid}, wantStatus: http.StatusOK, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// During key rotation Orb may send several candidates. Authentication
			// succeeds only if at least one complete v1 HMAC matches.
			jobClient := &orbWebhookJobs{}
			response := serveOrbWebhook(t, jobClient, payload, tt.timestamp, tt.signatures...)

			require.Equal(t, tt.wantStatus, response.Code)
			require.Equal(t, tt.wantCalls, jobClient.callCount())
			if tt.wantCalls == 1 {
				require.Equal(t, []orbPaymentSuccessCall{{
					billingCustomerID: "org_invoice_customer",
					invoiceID:         "inv_bill_001",
				}}, jobClient.paymentSuccessCalls)
			}
		})
	}
}

func TestOrbWebhookSignatureTimestampWindow(t *testing.T) {
	// Pin the inclusive replay window at both five-minute boundaries without relying on wall-clock timing.
	now := time.Date(2026, time.July, 23, 9, 45, 0, 123456789, time.UTC)
	payload := orbInvoicePayload(t, "invoice.payment_succeeded")
	verifier := &orbWebhook{orb: &Orb{webhookSecret: testOrbWebhookSecret}}

	tests := []struct {
		name        string
		timestamp   time.Time
		wantErrText string
	}{
		{name: "exactly five minutes old", timestamp: now.Add(-5 * time.Minute)},
		{name: "one nanosecond too old", timestamp: now.Add(-5*time.Minute - time.Nanosecond), wantErrText: "too old"},
		{name: "current", timestamp: now},
		{name: "exactly five minutes new", timestamp: now.Add(5 * time.Minute)},
		{name: "one nanosecond too new", timestamp: now.Add(5*time.Minute + time.Nanosecond), wantErrText: "too new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The freshness window is inclusive at both five-minute boundaries;
			// anything outside it is rejected even when its HMAC is otherwise valid.
			timestamp := tt.timestamp.Format(webhookHeaderTimestampFormat)
			headers := make(http.Header)
			headers.Set("X-Orb-Timestamp", timestamp)
			headers.Set("X-Orb-Signature", orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

			err := verifier.verifySignature(payload, headers, now)
			if tt.wantErrText == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErrText)
			}
		})
	}
}

func TestOrbWebhookReplayUsesDurableQueueDeduplication(t *testing.T) {
	// Webhook delivery is at-least-once. Both deliveries must reach the durable
	// queue with identical keys; its Duplicate result makes the replay a safe 200.
	payload := orbInvoicePayload(t, "invoice.payment_succeeded")
	timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
	signature := orbSignedHeader(payload, timestamp, testOrbWebhookSecret)
	jobClient := &orbWebhookJobs{duplicateOnRepeat: true}

	first := serveOrbWebhook(t, jobClient, payload, timestamp, signature)
	second := serveOrbWebhook(t, jobClient, payload, timestamp, signature)

	require.Equal(t, http.StatusOK, first.Code)
	require.Equal(t, http.StatusOK, second.Code)
	require.Equal(t, []orbPaymentSuccessCall{
		{billingCustomerID: "org_invoice_customer", invoiceID: "inv_bill_001"},
		{billingCustomerID: "org_invoice_customer", invoiceID: "inv_bill_001"},
	}, jobClient.paymentSuccessCalls)
	require.Equal(t, 2, jobClient.callCount())
}

func TestOrbWebhookBodyValidation(t *testing.T) {
	// Signed input is still untrusted: malformed events and bodies beyond the explicit limit must fail before enqueue.
	t.Run("malformed JSON", func(t *testing.T) {
		payload := []byte(`{"type":`)
		timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
		jobClient := &orbWebhookJobs{}
		response := serveOrbWebhook(t, jobClient, payload, timestamp, orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, jobClient.callCount())
	})

	t.Run("malformed recognized event", func(t *testing.T) {
		payload := []byte(`{"type":"invoice.payment_succeeded","created_at":42}`)
		timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
		jobClient := &orbWebhookJobs{}
		response := serveOrbWebhook(t, jobClient, payload, timestamp, orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, jobClient.callCount())
	})

	t.Run("exactly 64 KiB", func(t *testing.T) {
		payload := []byte(`{"type":"event.not_used"}`)
		payload = append(payload, bytes.Repeat([]byte(" "), int(maxBodyBytes)-len(payload))...)
		jobClient := &orbWebhookJobs{}
		response := serveOrbWebhook(t, jobClient, payload, "")

		require.Equal(t, int(maxBodyBytes), len(payload))
		require.Equal(t, http.StatusOK, response.Code)
		require.Zero(t, jobClient.callCount())
	})

	t.Run("larger than 64 KiB", func(t *testing.T) {
		payload := bytes.Repeat([]byte("x"), int(maxBodyBytes)+1)
		timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
		jobClient := &orbWebhookJobs{}
		response := serveOrbWebhook(t, jobClient, payload, timestamp, orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
		require.Zero(t, jobClient.callCount())
	})
}

func TestOrbWebhookQueueFailuresReturnRetryableStatus(t *testing.T) {
	// Every supported event returns a retryable server error when its handoff to the durable queue fails.
	queueErr := errors.New("queue unavailable")
	tests := []struct {
		name       string
		eventType  string
		payload    func(*testing.T, string) []byte
		assertJobs func(*testing.T, *orbWebhookJobs)
	}{
		{
			name: "payment success", eventType: "invoice.payment_succeeded", payload: orbInvoicePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []orbPaymentSuccessCall{{billingCustomerID: "org_invoice_customer", invoiceID: "inv_bill_001"}}, got.paymentSuccessCalls)
			},
		},
		{
			name: "payment failed", eventType: "invoice.payment_failed", payload: orbInvoicePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []orbPaymentFailedCall{{
					billingCustomerID: "org_invoice_customer",
					invoiceID:         "inv_bill_001",
					invoiceNumber:     "RILL-000042",
					invoiceURL:        "https://billing.example.test/invoices/inv_bill_001",
					amount:            "42.75",
					currency:          "USD",
					dueDate:           testOrbDueDate,
					failedAt:          testOrbPaymentFailed,
				}}, got.paymentFailedCalls)
			},
		},
		{
			name: "plan changed", eventType: "subscription.plan_changed", payload: orbSubscriptionPayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_subscription_customer"}, got.planChangedCalls)
			},
		},
		{
			name: "credit dropped", eventType: "customer.credit_balance_dropped", payload: orbCreditBalancePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_credit_customer"}, got.creditBalanceDroppedCalls)
			},
		},
		{
			name: "credit depleted", eventType: "customer.credit_balance_depleted", payload: orbCreditBalancePayload,
			assertJobs: func(t *testing.T, got *orbWebhookJobs) {
				require.Equal(t, []string{"org_credit_customer"}, got.creditBalanceDepletedCalls)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A failed insert must never be acknowledged: the 5xx is what asks Orb
			// to retry, while the recorded call proves no identifiers were dropped.
			payload := tt.payload(t, tt.eventType)
			timestamp := time.Now().UTC().Format(webhookHeaderTimestampFormat)
			jobClient := &orbWebhookJobs{queueErr: queueErr}
			response := serveOrbWebhook(t, jobClient, payload, timestamp, orbSignedHeader(payload, timestamp, testOrbWebhookSecret))

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.Equal(t, 1, jobClient.callCount())
			tt.assertJobs(t, jobClient)
		})
	}
}

func orbInvoicePayload(t *testing.T, eventType string) []byte {
	t.Helper()
	return orbJSONPayload(t, map[string]any{
		"id":         "evt_invoice_bill_001",
		"type":       eventType,
		"created_at": "2026-07-23T08:00:00Z",
		"properties": map[string]any{"reason": "test"},
		"invoice": map[string]any{
			"id":                 "inv_bill_001",
			"invoice_number":     "RILL-000042",
			"hosted_invoice_url": "https://billing.example.test/invoices/inv_bill_001",
			"amount_due":         "42.75",
			"currency":           "USD",
			"due_date":           testOrbDueDate.Format(time.RFC3339Nano),
			"payment_failed_at":  testOrbPaymentFailed.Format(time.RFC3339Nano),
			"customer": map[string]any{
				"id":                   "orb_customer_invoice",
				"external_customer_id": "org_invoice_customer",
			},
		},
	})
}

func orbSubscriptionPayload(t *testing.T, eventType string) []byte {
	t.Helper()
	return orbJSONPayload(t, map[string]any{
		"id":         "evt_subscription_bill_001",
		"type":       eventType,
		"created_at": "2026-07-23T08:00:00Z",
		"subscription": map[string]any{
			"id": "sub_bill_001",
			"customer": map[string]any{
				"id":                   "orb_customer_subscription",
				"external_customer_id": "org_subscription_customer",
			},
		},
	})
}

func orbCreditBalancePayload(t *testing.T, eventType string) []byte {
	t.Helper()
	return orbJSONPayload(t, map[string]any{
		"id":         "evt_credit_bill_001",
		"type":       eventType,
		"created_at": "2026-07-23T08:00:00Z",
		"customer": map[string]any{
			"id":                   "orb_customer_credit",
			"external_customer_id": "org_credit_customer",
		},
	})
}

func orbJSONPayload(t *testing.T, event map[string]any) []byte {
	t.Helper()
	payload, err := json.Marshal(event)
	require.NoError(t, err)
	return payload
}

func orbSignedHeader(payload []byte, timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("v1:"))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(":"))
	mac.Write(payload)
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func serveOrbWebhook(t *testing.T, jobClient jobs.Client, payload []byte, timestamp string, signatures ...string) *httptest.ResponseRecorder {
	t.Helper()
	handler := NewOrb(zap.NewNop(), "", testOrbWebhookSecret, "").WebhookHandlerFunc(context.Background(), jobClient)
	require.NotNil(t, handler)

	request := httptest.NewRequest(http.MethodPost, "/orb/webhook", bytes.NewReader(payload))
	if timestamp != "" {
		request.Header.Set("X-Orb-Timestamp", timestamp)
	}
	for _, signature := range signatures {
		request.Header.Add("X-Orb-Signature", signature)
	}
	response := httptest.NewRecorder()
	require.NotPanics(t, func() {
		handler.ServeHTTP(response, request)
	})
	return response
}
