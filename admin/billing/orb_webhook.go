package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/orbcorp/orb-go"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
	"github.com/riverqueue/river"
)

// WebhookHeaderTimestampFormat is the format of the header X-Orb-Timestamp for webhook requests sent by Orb.
const WebhookHeaderTimestampFormat = "2006-01-02T15:04:05.999999999"

var interestingEvents = []string{"invoice.payment_succeeded", "invoice.payment_failed"}

func (o *Orb) handleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusServiceUnavailable)
		return
	}

	// unmarshal event first before even verifying signature, so we can ignore unwanted events as Orb does not have an option to selectively deliver only required events
	var e genericEvent
	err = json.Unmarshal(payload, &e)
	if err != nil {
		http.Error(w, "error parsing event data", http.StatusBadRequest)
		return
	}

	if !slices.Contains(interestingEvents, e.Type) {
		w.WriteHeader(http.StatusOK)
		return
	}

	now := time.Now().UTC()
	err = o.verifySignature(payload, r.Header, now)
	if err != nil {
		http.Error(w, "error verifying webhook signature", http.StatusBadRequest)
		return
	}

	switch e.Type {
	case "invoice.payment_succeeded":
		var ie invoiceEvent
		err = json.Unmarshal(payload, &ie)
		if err != nil {
			http.Error(w, "error parsing event data", http.StatusBadRequest)
			return
		}
		err = o.handleInvoicePaymentSucceeded(r.Context(), ie)
		if err != nil {
			http.Error(w, "error handling event", http.StatusInternalServerError)
			return
		}
	case "invoice.payment_failed":
		var ie invoiceEvent
		err = json.Unmarshal(payload, &ie)
		if err != nil {
			http.Error(w, "error parsing event data", http.StatusBadRequest)
			return
		}
		err = o.handleInvoicePaymentFailed(r.Context(), ie)
		if err != nil {
			http.Error(w, "error handling event", http.StatusInternalServerError)
			return
		}
	default:
		// do nothing
	}

	w.WriteHeader(http.StatusOK)
}

func (o *Orb) handleInvoicePaymentSucceeded(ctx context.Context, ie invoiceEvent) error {
	_, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.InvoicePaymentSuccessArgs{
		BillingCustomerID: ie.OrbInvoice.Customer.ExternalCustomerID,
		InvoiceID:         ie.OrbInvoice.ID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	return err
}

func (o *Orb) handleInvoicePaymentFailed(ctx context.Context, ie invoiceEvent) error {
	_, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.InvoicePaymentFailedArgs{
		BillingCustomerID: ie.OrbInvoice.Customer.ExternalCustomerID,
		InvoiceID:         ie.OrbInvoice.ID,
		InvoiceNumber:     ie.OrbInvoice.InvoiceNumber,
		InvoiceURL:        ie.OrbInvoice.HostedInvoiceURL,
		Amount:            ie.OrbInvoice.AmountDue,
		Currency:          ie.OrbInvoice.Currency,
		DueDate:           ie.OrbInvoice.DueDate,
		FailedAt:          ie.OrbInvoice.PaymentFailedAt,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	return err
}

// Validates whether or not the webhook payload was sent by Orb.
func (o *Orb) verifySignature(payload []byte, headers http.Header, now time.Time) error {
	if o.webhookSecret == "" {
		return errors.New("no webhook secret set")
	}

	msgSignature := headers.Values("X-Orb-Signature")
	if len(msgSignature) == 0 {
		return errors.New("could not find X-Orb-Signature header")
	}
	msgTimestamp := headers.Get("X-Orb-Timestamp")
	if msgTimestamp == "" {
		return errors.New("could not find X-Orb-Timestamp header")
	}

	timestamp, err := time.Parse(WebhookHeaderTimestampFormat, msgTimestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp headers: %w", err)
	}

	if timestamp.Before(now.Add(-5 * time.Minute)) {
		return errors.New("value from X-Orb-Timestamp header too old")
	}
	if timestamp.After(now.Add(5 * time.Minute)) {
		return errors.New("value from X-Orb-Timestamp header too new")
	}

	secretBytes := []byte(o.webhookSecret)
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte("v1:"))
	mac.Write([]byte(msgTimestamp))
	mac.Write([]byte(":"))
	mac.Write(payload)
	expected := mac.Sum(nil)

	for _, part := range msgSignature {
		parts := strings.Split(part, "=")
		if len(parts) != 2 {
			continue
		}
		if parts[0] != "v1" {
			continue
		}
		signature, err := hex.DecodeString(parts[1])
		if err != nil {
			continue
		}
		if hmac.Equal(signature, expected) {
			return nil
		}
	}

	return errors.New("none of the given webhook signatures match the expected signature")
}

// just to unmarshal the event type
type genericEvent struct {
	Type string `json:"type"`
}

type invoiceEvent struct {
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	OrbInvoice orb.Invoice `json:"invoice"`
}
