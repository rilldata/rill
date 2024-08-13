package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
	"github.com/riverqueue/river"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
)

// handleWebhook handles incoming webhook events from Stripe
func (s *Stripe) handleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	endpointSecret := s.webhookSecret
	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		s.logger.Error("Error verifying webhook signature", zap.Error(err))
		http.Error(w, "Error verifying webhook signature", http.StatusBadRequest)
		return
	}

	s.logger.Debug(fmt.Sprintf("got event : %s\n", event.Data))

	// Handle the event based on its type
	switch event.Type {
	case "charge.succeeded":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			s.logger.Error(fmt.Sprintf("Error parsing charge data: %v", err))
			http.Error(w, "Error parsing charge data", http.StatusBadRequest)
			return
		}
		charge.Customer = &stripe.Customer{ID: "cus_QeIXiyS2fAGoDx"}
		if charge.Customer == nil {
			s.logger.Error(fmt.Sprintf("No customer info sent for charge %s", charge.ID))
			http.Error(w, "Error parsing charge data", http.StatusBadRequest)
			return
		}
		err = s.handleChargeSucceeded(r.Context(), &charge)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error handling charge.succeeded event: %v", err))
			http.Error(w, "Error handling charge.succeeded event", http.StatusInternalServerError)
			return
		}
	case "charge.failed":
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			s.logger.Error(fmt.Sprintf("Error parsing charge data: %v", err))
			http.Error(w, "Error parsing charge data", http.StatusBadRequest)
			return
		}
		charge.Customer = &stripe.Customer{ID: "cus_QeIXiyS2fAGoDx"}
		if charge.Customer == nil {
			s.logger.Error(fmt.Sprintf("No customer info sent for charge %s", charge.ID))
			http.Error(w, "Error parsing charge data", http.StatusBadRequest)
			return
		}
		err = s.handleChargeFailed(r.Context(), &charge)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error handling charge.failed event: %v", err))
			http.Error(w, "Error handling charge.failed event", http.StatusInternalServerError)
			return
		}
	default:
		s.logger.Warn(fmt.Sprintf("Unhandled event type: %s\n", event.Type))
	}

	// Acknowledge receipt of the event
	w.WriteHeader(http.StatusOK)
}

func (s *Stripe) handleChargeSucceeded(ctx context.Context, charge *stripe.Charge) error {
	metadata := map[string]string{"charge_id": charge.ID}
	if charge.Amount != 0 { // being defensive
		metadata["amount"] = fmt.Sprintf("%s %d", charge.Currency, charge.Amount)
	}
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.ChargeSuccessArgs{
		CustomerID: charge.Customer.ID,
		Metadata:   metadata,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add charge success: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		// TODO change to debug
		s.logger.Info("Duplicate charge success event", zap.String("customer_id", charge.Customer.ID), zap.String("customer_name", charge.Customer.Name), zap.String("customer_email", charge.Customer.Email))
		return nil
	}
	return nil
}

func (s *Stripe) handleChargeFailed(ctx context.Context, charge *stripe.Charge) error {
	metadata := map[string]string{"charge_id": charge.ID}
	if charge.Amount != 0 { // being defensive
		metadata["amount"] = fmt.Sprintf("%s %d", charge.Currency, charge.Amount)
	}
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.AddBillingErrorArgs{
		CustomerID: charge.Customer.ID,
		ErrorType:  database.BillingErrorTypePaymentFailed,
		Metadata:   metadata,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		// TODO change to debug
		s.logger.Info("Duplicate billing error event", zap.String("customer_id", charge.Customer.ID), zap.String("customer_name", charge.Customer.Name), zap.String("customer_email", charge.Customer.Email))
		return nil
	}
	return nil
}
