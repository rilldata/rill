package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var _ Client = &client{}

type client struct {
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// New creates a HubSpot client that upserts contacts via the HubSpot v3 API.
func New(logger *zap.Logger, apiKey string) Client {
	return &client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.Named("hubspot"),
	}
}

// UpsertContact creates or updates a contact in HubSpot by email.
// The call is non-blocking; it runs in a background goroutine with panic recovery.
func (c *client) UpsertContact(email string, properties map[string]string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("panic in HubSpot upsert", zap.String("email", email), zap.Any("recover", r))
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := c.upsertContact(ctx, email, properties)
		if err != nil {
			c.logger.Error("failed to upsert HubSpot contact", zap.String("email", email), zap.Error(err))
		}
	}()
}

// upsertContact performs the actual HubSpot API call.
// Uses the "create or update" endpoint: POST /crm/v3/objects/contacts with idProperty=email.
func (c *client) upsertContact(ctx context.Context, email string, properties map[string]string) error {
	// Add email to properties
	props := make(map[string]string, len(properties)+1)
	for k, v := range properties {
		props[k] = v
	}
	props["email"] = email

	body := map[string]interface{}{
		"properties": props,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Try to create the contact first
	err = c.createContact(ctx, bodyBytes)
	if err == nil {
		return nil
	}

	// If conflict (contact exists), update by email
	if isConflictError(err) {
		return c.updateContactByEmail(ctx, email, props)
	}

	return err
}

func (c *client) createContact(ctx context.Context, bodyBytes []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.hubapi.com/crm/v3/objects/contacts", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusBadRequest {
		// 409 Conflict or 400 with "contact already exists" means we need to update
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return &conflictError{status: resp.StatusCode, body: errResp}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("HubSpot API returned %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

func (c *client) updateContactByEmail(ctx context.Context, email string, properties map[string]string) error {
	body := map[string]interface{}{
		"properties": properties,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal update request: %w", err)
	}

	// PATCH /crm/v3/objects/contacts/{email}?idProperty=email
	url := fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/contacts/%s?idProperty=email", email)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("HubSpot update API returned %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

type conflictError struct {
	status int
	body   map[string]interface{}
}

func (e *conflictError) Error() string {
	return fmt.Sprintf("HubSpot conflict (status %d): %v", e.status, e.body)
}

func isConflictError(err error) bool {
	_, ok := err.(*conflictError)
	return ok
}
