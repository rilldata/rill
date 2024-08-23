package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
	case stripe.EventType(database.StripeWebhookEventTypePaymentMethodAttached):
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			s.logger.Error(fmt.Sprintf("Error parsing payment method data: %v", err))
			http.Error(w, "Error parsing payment method data", http.StatusBadRequest)
			return
		}
		if paymentMethod.Customer == nil {
			s.logger.Error(fmt.Sprintf("No customer info sent for payment method %s", paymentMethod.ID))
			http.Error(w, "Error parsing payment method data", http.StatusBadRequest)
			return
		}
		err = s.handlePaymentMethodAdded(r.Context(), &paymentMethod)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error handling payment_method.attached event: %v", err))
			http.Error(w, "Error handling payment_method.attached event", http.StatusInternalServerError)
			return
		}
	case stripe.EventType(database.StripeWebhookEventTypePaymentMethodDetached):
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			s.logger.Error(fmt.Sprintf("Error parsing payment method data: %v", err))
			http.Error(w, "Error parsing payment method data", http.StatusBadRequest)
			return
		}
		if paymentMethod.Customer == nil {
			s.logger.Error(fmt.Sprintf("No customer info sent for payment method %s", paymentMethod.ID))
			http.Error(w, "Error parsing payment method data", http.StatusBadRequest)
			return
		}
		err = s.handlePaymentMethodRemoved(r.Context(), &paymentMethod)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error handling payment_method.detached event: %v", err))
			http.Error(w, "Error handling payment_method.detached event", http.StatusInternalServerError)
			return
		}
	default:
		s.logger.Warn(fmt.Sprintf("Unhandled event type: %s\n", event.Type))
	}

	// Acknowledge receipt of the event
	w.WriteHeader(http.StatusOK)
}

func (s *Stripe) handlePaymentMethodAdded(ctx context.Context, method *stripe.PaymentMethod) error {
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.PaymentMethodAddedArgs{
		ID:                method.ID,
		PaymentCustomerID: method.Customer.ID,
		PaymentType:       string(method.Type),
		EventTime:         time.UnixMilli(method.Created),
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add payment method added: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		s.logger.Debug("Duplicate payment method added event", zap.String("customer_id", method.Customer.ID), zap.String("customer_name", method.Customer.Name), zap.String("customer_email", method.Customer.Email))
		return nil
	}
	return nil
}

func (s *Stripe) handlePaymentMethodRemoved(ctx context.Context, method *stripe.PaymentMethod) error {
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.PaymentMethodRemovedArgs{
		ID:                method.ID,
		PaymentCustomerID: method.Customer.ID,
		EventTime:         time.UnixMilli(method.Created),
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add payment method added: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		s.logger.Debug("Duplicate payment method added event", zap.String("customer_id", method.Customer.ID), zap.String("customer_name", method.Customer.Name), zap.String("customer_email", method.Customer.Email))
		return nil
	}
	return nil
}
