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
		s.logger.Error("error reading request body", zap.Error(err))
		http.Error(w, "error reading request body", http.StatusServiceUnavailable)
		return
	}

	endpointSecret := s.webhookSecret
	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(payload, sigHeader, endpointSecret, webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		s.logger.Error("error verifying webhook signature", zap.Error(err))
		http.Error(w, "error verifying webhook signature", http.StatusBadRequest)
		return
	}

	s.logger.Debug(fmt.Sprintf("got event : %s\n", event.Data))

	// Handle the event based on its type
	switch event.Type {
	case stripe.EventType(database.StripeWebhookEventTypePaymentMethodAttached):
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			s.logger.Error("error parsing payment method data", zap.Error(err))
			http.Error(w, "error parsing payment method data", http.StatusBadRequest)
			return
		}
		if paymentMethod.Customer == nil {
			// just log warn and send http ok as we can't do anything without customer id
			s.logger.Warn("no customer info sent for payment_method.attached event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		} else {
			err = s.handlePaymentMethodAdded(r.Context(), &paymentMethod)
			if err != nil {
				s.logger.Error("Error handling payment_method.attached event", zap.Error(err))
				http.Error(w, "Error handling payment_method.attached event", http.StatusInternalServerError)
				return
			}
		}
	case stripe.EventType(database.StripeWebhookEventTypePaymentMethodDetached):
		var paymentMethod stripe.PaymentMethod
		if err := json.Unmarshal(event.Data.Raw, &paymentMethod); err != nil {
			s.logger.Error("Error parsing payment method data", zap.Error(err))
			http.Error(w, "Error parsing payment method data", http.StatusBadRequest)
			return
		}
		if cust, ok := event.Data.PreviousAttributes["customer"]; ok && cust != nil {
			err = s.handlePaymentMethodRemoved(r.Context(), cust.(string), &paymentMethod)
			if err != nil {
				s.logger.Error("error handling payment_method.detached event", zap.Error(err))
				http.Error(w, "error handling payment_method.detached event", http.StatusInternalServerError)
				return
			}
		} else {
			// just log warn and send http ok as we can't do anything without customer id
			s.logger.Warn("no customer info sent for payment method detached event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		}
	case stripe.EventType(database.StripeWebhookEventTypeCustomerUpdated):
		var customer stripe.Customer
		if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
			s.logger.Error("error parsing customer data", zap.Error(err))
			http.Error(w, "error parsing customer data", http.StatusBadRequest)
			return
		}
		if customer.ID == "" {
			// just log warn and send http ok as we can't do anything without customer id
			s.logger.Warn("no customer info sent for customer.updated event", zap.String("event_id", event.ID), zap.Time("event_time", time.UnixMilli(event.Created*1000)))
		} else {
			// we just care about address update, so check if address was update and now customer has a valid address
			if _, ok := event.Data.PreviousAttributes["address"]; ok && customer.Address != nil {
				err = s.handleCustomerAddressUpdated(r.Context(), &customer)
				if err != nil {
					s.logger.Error("error handling customer.updated event", zap.Error(err))
					http.Error(w, "error handling customer.updated event", http.StatusInternalServerError)
					return
				}
			}
		}
	default:
		s.logger.Warn("unhandled stripe event type", zap.String("type", string(event.Type)))
	}

	// Acknowledge receipt of the event
	w.WriteHeader(http.StatusOK)
}

func (s *Stripe) handlePaymentMethodAdded(ctx context.Context, method *stripe.PaymentMethod) error {
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.PaymentMethodAddedArgs{
		PaymentMethodID:   method.ID,
		PaymentCustomerID: method.Customer.ID,
		PaymentType:       string(method.Type),
		EventTime:         time.UnixMilli(method.Created * 1000),
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add payment method added: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		s.logger.Debug("duplicate payment method added event", zap.String("customer_id", method.Customer.ID))
		return nil
	}
	return nil
}

func (s *Stripe) handlePaymentMethodRemoved(ctx context.Context, customerID string, method *stripe.PaymentMethod) error {
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.PaymentMethodRemovedArgs{
		PaymentMethodID:   method.ID,
		PaymentCustomerID: customerID,
		EventTime:         time.UnixMilli(method.Created * 1000),
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add payment method added: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		s.logger.Debug("duplicate payment method added event", zap.String("customer_id", customerID))
		return nil
	}
	return nil
}

func (s *Stripe) handleCustomerAddressUpdated(ctx context.Context, customer *stripe.Customer) error {
	res, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.CustomerAddressUpdatedArgs{
		PaymentCustomerID: customer.ID,
		EventTime:         time.UnixMilli(customer.Created * 1000),
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add customer updated event: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		s.logger.Debug("duplicate customer updated event", zap.String("customer_id", customer.ID))
		return nil
	}
	return nil
}
