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
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"go.uber.org/zap"
)

const (
	webhookHeaderTimestampFormat = "2006-01-02T15:04:05.999999999" // format of the header X-Orb-Timestamp for webhook requests sent by Orb.
	maxBodyBytes                 = int64(65536)
)

var interestingEvents = []string{"invoice.payment_succeeded", "invoice.payment_failed", "invoice.issue_failed"}

type orbWebhook struct {
	orb  *Orb
	jobs jobs.Client
}

func (o *orbWebhook) handleWebhook(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return httputil.Errorf(http.StatusServiceUnavailable, "error reading request body: %w", err)
	}

	// unmarshal event first before even verifying signature, so we can ignore unwanted events as Orb does not have an option to selectively deliver only required events
	var e genericEvent
	err = json.Unmarshal(payload, &e)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "error parsing event data: %w", err)
	}

	if !slices.Contains(interestingEvents, e.Type) {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	now := time.Now().UTC()
	err = o.verifySignature(payload, r.Header, now)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "error verifying webhook signature: %w", err)
	}

	switch e.Type {
	case "invoice.payment_succeeded":
		var ie invoiceEvent
		err = json.Unmarshal(payload, &ie)
		if err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing event data: %w", err)
		}
		err = o.handleInvoicePaymentSucceeded(r.Context(), ie)
		if err != nil {
			return httputil.Errorf(http.StatusInternalServerError, "error handling event: %w", err)
		}
	case "invoice.payment_failed":
		var ie invoiceEvent
		err = json.Unmarshal(payload, &ie)
		if err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing event data: %w", err)
		}
		err = o.handleInvoicePaymentFailed(r.Context(), ie)
		if err != nil {
			return httputil.Errorf(http.StatusInternalServerError, "error handling event: %w", err)
		}
	case "invoice.issue_failed":
		var ie invoiceEvent
		err = json.Unmarshal(payload, &ie)
		if err != nil {
			return httputil.Errorf(http.StatusBadRequest, "error parsing event data: %w", err)
		}
		// inefficient one time conversion to named logger as its rare event and no need to log every thing else with named logger
		o.orb.logger.Named("billing").Warn("invoice issue failed", zap.String("customer_id", ie.OrbInvoice.Customer.ExternalCustomerID), zap.String("invoice_id", ie.OrbInvoice.ID), zap.String("props", fmt.Sprintf("%v", ie.Properties)))
	default:
		// do nothing
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (o *orbWebhook) handleInvoicePaymentSucceeded(ctx context.Context, ie invoiceEvent) error {
	res, err := o.jobs.PaymentSuccess(ctx, ie.OrbInvoice.Customer.ExternalCustomerID, ie.OrbInvoice.ID)
	if err != nil {
		return err
	}
	if res.Duplicate {
		o.orb.logger.Debug("duplicate invoice payment success event", zap.String("event_d", ie.ID))
	}
	return nil
}

func (o *orbWebhook) handleInvoicePaymentFailed(ctx context.Context, ie invoiceEvent) error {
	res, err := o.jobs.PaymentFailed(ctx,
		ie.OrbInvoice.Customer.ExternalCustomerID,
		ie.OrbInvoice.ID,
		ie.OrbInvoice.InvoiceNumber,
		ie.OrbInvoice.HostedInvoiceURL,
		ie.OrbInvoice.AmountDue,
		ie.OrbInvoice.Currency,
		ie.OrbInvoice.DueDate,
		ie.OrbInvoice.PaymentFailedAt,
	)
	if err != nil {
		return err
	}
	if res.Duplicate {
		o.orb.logger.Debug("duplicate invoice payment failed event", zap.String("event_id", ie.ID))
	}
	return nil
}

// Validates whether or not the webhook payload was sent by Orb.
func (o *orbWebhook) verifySignature(payload []byte, headers http.Header, now time.Time) error {
	if o.orb.webhookSecret == "" {
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

	timestamp, err := time.Parse(webhookHeaderTimestampFormat, msgTimestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp headers: %w", err)
	}

	if timestamp.Before(now.Add(-5 * time.Minute)) {
		return errors.New("value from X-Orb-Timestamp header too old")
	}
	if timestamp.After(now.Add(5 * time.Minute)) {
		return errors.New("value from X-Orb-Timestamp header too new")
	}

	secretBytes := []byte(o.orb.webhookSecret)
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
