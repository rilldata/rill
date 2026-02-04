// Script to update all Orb subscriptions with the ACH/Wire instructions memo.
//
// Usage:
//
//	ORB_API_KEY=<your-orb-api-key> go run ./scripts/update_orb_subscription_memos
//
// Optional flags:
//
//	-dry-run      Print what would be updated without making changes
//	-status       Filter by subscription status (default: "active")
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/orbcorp/orb-go"
	"github.com/orbcorp/orb-go/option"
)

const (
	// The new ACH/Wire instructions memo to add to all subscriptions
	invoiceMemo = `ACH/Wire Instructions:
JPMorgan Chase New York, NY 10017
Account Name: Rill Data, Inc.
Account Number: 80008318778
Bank ABA/Routing: 321081669
SWIFT: CHASUS33`

	paginationLimit = 100
	requestTimeout  = 30 * time.Second
)

func main() {
	// Parse command-line flags
	dryRun := flag.Bool("dry-run", false, "Print what would be updated without making changes")
	status := flag.String("status", "active", "Filter by subscription status (active, ended, upcoming)")
	flag.Parse()

	// Get API key from environment
	apiKey := os.Getenv("ORB_API_KEY")
	if apiKey == "" {
		log.Fatal("ORB_API_KEY environment variable is required")
	}

	// Create Orb client
	client := orb.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(requestTimeout),
	)

	ctx := context.Background()

	// Convert status string to Orb status type
	var orbStatus orb.SubscriptionListParamsStatus
	switch *status {
	case "active":
		orbStatus = orb.SubscriptionListParamsStatusActive
	case "ended":
		orbStatus = orb.SubscriptionListParamsStatusEnded
	case "upcoming":
		orbStatus = orb.SubscriptionListParamsStatusUpcoming
	default:
		log.Fatalf("Invalid status: %s. Must be one of: active, ended, upcoming", *status)
	}

	// Fetch all subscriptions using pagination
	subscriptions, err := listAllSubscriptions(ctx, client, orbStatus)
	if err != nil {
		log.Fatalf("Failed to list subscriptions: %v", err)
	}

	fmt.Printf("Found %d subscriptions with status '%s'\n", len(subscriptions), *status)

	if len(subscriptions) == 0 {
		fmt.Println("No subscriptions to update.")
		return
	}

	// Update each subscription with the new memo
	successCount := 0
	errorCount := 0

	for i, sub := range subscriptions {
		customerName := sub.Customer.Name
		if customerName == "" {
			customerName = sub.Customer.ExternalCustomerID
		}

		fmt.Printf("[%d/%d] Processing subscription %s (Customer: %s)\n",
			i+1, len(subscriptions), sub.ID, customerName)

		if *dryRun {
			fmt.Printf("  [DRY-RUN] Would update subscription %s with new memo\n", sub.ID)
			successCount++
			continue
		}

		// Update the subscription with the new memo
		_, err := client.Subscriptions.Update(ctx, sub.ID, orb.SubscriptionUpdateParams{
			DefaultInvoiceMemo: orb.String(invoiceMemo),
		})
		if err != nil {
			fmt.Printf("  ERROR: Failed to update subscription %s: %v\n", sub.ID, err)
			errorCount++
			continue
		}

		fmt.Printf("  SUCCESS: Updated subscription %s\n", sub.ID)
		successCount++
	}

	// Print summary
	fmt.Println("\n--- Summary ---")
	fmt.Printf("Total subscriptions: %d\n", len(subscriptions))
	fmt.Printf("Successfully updated: %d\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)

	if *dryRun {
		fmt.Println("\nThis was a dry run. No changes were made.")
		fmt.Println("Run without -dry-run to apply changes.")
	}
}

// listAllSubscriptions fetches all subscriptions with the given status using pagination
func listAllSubscriptions(ctx context.Context, client *orb.Client, status orb.SubscriptionListParamsStatus) ([]orb.Subscription, error) {
	var allSubscriptions []orb.Subscription
	var cursor string

	for {
		params := orb.SubscriptionListParams{
			Status: orb.F(status),
			Limit:  orb.Int(paginationLimit),
		}

		if cursor != "" {
			params.Cursor = orb.String(cursor)
		}

		page, err := client.Subscriptions.List(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list subscriptions: %w", err)
		}

		allSubscriptions = append(allSubscriptions, page.Data...)

		// Check if there are more pages
		if page.PaginationMetadata.HasMore {
			cursor = page.PaginationMetadata.NextCursor
		} else {
			break
		}
	}

	return allSubscriptions, nil
}
