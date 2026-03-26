package hubspot

// Client provides methods for syncing user and organization data to HubSpot.
// It follows the same interface + noop pattern as admin/billing.
type Client interface {
	// UpsertContact creates or updates a HubSpot contact by email.
	// Properties include name, org, role, trial dates, etc.
	// Implementations must be safe to call from request handlers (non-blocking).
	UpsertContact(email string, properties map[string]string)
}
