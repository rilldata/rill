package admin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

// TestStartTrialKeepsConcurrentQuotaUpdate checks that starting a trial does not revert a quota update
// that lands while the trial job is talking to the billing system.
// The trial jobs read the org before making their billing calls, so they must not write that snapshot back.
func TestStartTrialKeepsConcurrentQuotaUpdate(t *testing.T) {
	pg := pgtestcontainer.New(t)
	t.Cleanup(func() { pg.Terminate(t) })

	keyring, err := database.NewRandomKeyring()
	require.NoError(t, err)
	keyringJSON, err := json.Marshal(keyring)
	require.NoError(t, err)
	db, err := database.Open("postgres", pg.DatabaseURL, string(keyringJSON))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Migrate(t.Context()))

	tests := []struct {
		name    string
		orgName string
		start   func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, *billing.Subscription, error)
	}{
		{name: "credit trial", orgName: "credit-trial", start: (*Service).StartCreditTrial},
		{name: "legacy trial", orgName: "legacy-trial", start: (*Service).StartTrial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{
				Name:                    tt.orgName,
				QuotaProjects:           1,
				QuotaDeployments:        1,
				QuotaSlotsTotal:         1,
				QuotaSlotsPerDeployment: 1,
				QuotaOutstandingInvites: 1,
				QuotaSeats:              1,
				BillingCustomerID:       tt.orgName,
			})
			require.NoError(t, err)

			biller := &quotaRaceBiller{
				Biller: billing.NewNoop(),
				// Simulates `rill sudo quota set --projects 10` landing mid-flight.
				onCreateSubscription: func(ctx context.Context) {
					latest, err := db.FindOrganization(ctx, org.ID)
					require.NoError(t, err)
					_, err = db.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
						Name:                                latest.Name,
						DisplayName:                         latest.DisplayName,
						DefaultProvisioner:                  latest.DefaultProvisioner,
						QuotaProjects:                       10,
						QuotaDeployments:                    latest.QuotaDeployments,
						QuotaSlotsTotal:                     latest.QuotaSlotsTotal,
						QuotaSlotsPerDeployment:             latest.QuotaSlotsPerDeployment,
						QuotaOutstandingInvites:             latest.QuotaOutstandingInvites,
						QuotaStorageLimitBytesPerDeployment: latest.QuotaStorageLimitBytesPerDeployment,
						QuotaSeats:                          latest.QuotaSeats,
						BillingCustomerID:                   latest.BillingCustomerID,
						PaymentCustomerID:                   latest.PaymentCustomerID,
						BillingEmail:                        latest.BillingEmail,
						BillingPlanName:                     latest.BillingPlanName,
						BillingPlanDisplayName:              latest.BillingPlanDisplayName,
						CreatedByUserID:                     latest.CreatedByUserID,
					})
					require.NoError(t, err)
				},
			}
			svc := &Service{DB: db, Biller: biller, Logger: zap.NewNop()}

			// Pass the snapshot read before the quota update, like the river jobs do.
			updated, _, err := tt.start(svc, ctx, org)
			require.NoError(t, err)

			persisted, err := db.FindOrganization(ctx, org.ID)
			require.NoError(t, err)
			for _, o := range []*database.Organization{updated, persisted} {
				require.Equal(t, 10, o.QuotaProjects, "concurrent quota update was reverted")
				require.Equal(t, 5, o.QuotaSeats, "quota was not raised to the plan's quota")
				require.Equal(t, 1, o.QuotaDeployments)
				require.NotNil(t, o.BillingPlanName)
				require.Equal(t, "trial_plan", *o.BillingPlanName)
			}
		})
	}
}

// quotaRaceBiller is a biller with an active trial plan that runs onCreateSubscription while "creating" the subscription.
type quotaRaceBiller struct {
	billing.Biller
	onCreateSubscription func(ctx context.Context)
}

func (b *quotaRaceBiller) GetPlanByType(ctx context.Context, planType billing.PlanType) (*billing.Plan, error) {
	return &billing.Plan{
		ID:          "trial_plan_id",
		Name:        "trial_plan",
		PlanType:    planType,
		DisplayName: "Trial",
		Quotas: billing.Quotas{
			NumProjects: ptr(1),
			NumSeats:    ptr(5),
		},
	}, nil
}

func (b *quotaRaceBiller) GetActiveSubscription(ctx context.Context, customerID string) (*billing.Subscription, error) {
	return nil, billing.ErrNotFound
}

func (b *quotaRaceBiller) GetCustomerCreditBalance(ctx context.Context, customerID, currency string) (float64, error) {
	return billing.CreditTrialLowBalanceThreshold + 1, nil
}

func (b *quotaRaceBiller) CreateSubscription(ctx context.Context, customerID string, plan *billing.Plan) (*billing.Subscription, error) {
	b.onCreateSubscription(ctx)
	now := time.Now()
	return &billing.Subscription{ID: "sub", Plan: plan, StartDate: now, TrialEndDate: now.AddDate(0, 0, 30)}, nil
}

func ptr[T any](v T) *T {
	return &v
}
