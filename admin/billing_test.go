package admin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

// TestBillingKeepsConcurrentQuotaUpdate checks that billing flows do not revert a quota update
// that lands while they are talking to the billing system.
// They read the org before making their billing calls, so they must not write that snapshot back.
func TestBillingKeepsConcurrentQuotaUpdate(t *testing.T) {
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
		// activeSub makes the biller report an existing subscription on the trial plan.
		activeSub bool
		run       func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, error)
		// wantSeats is the plan's seat quota when the flow raises quotas to the plan.
		wantSeats    int
		wantPlanName *string
	}{
		{
			name:    "credit trial",
			orgName: "credit-trial",
			run: func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, error) {
				org, _, err := s.StartCreditTrial(ctx, org)
				return org, err
			},
			wantSeats:    5,
			wantPlanName: new("trial_plan"),
		},
		{
			name:    "legacy trial",
			orgName: "legacy-trial",
			run: func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, error) {
				org, _, err := s.StartTrial(ctx, org)
				return org, err
			},
			wantSeats:    5,
			wantPlanName: new("trial_plan"),
		},
		{
			name:    "init billing",
			orgName: "init-billing",
			run: func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, error) {
				return s.InitOrganizationBilling(ctx, org)
			},
			wantSeats: 1,
		},
		{
			name:      "repair billing with existing subscription",
			orgName:   "repair-billing",
			activeSub: true,
			run: func(s *Service, ctx context.Context, org *database.Organization) (*database.Organization, error) {
				org, _, err := s.RepairOrganizationBilling(ctx, org, true)
				return org, err
			},
			wantSeats: 5,
		},
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
				Biller:    billing.NewNoop(),
				activeSub: tt.activeSub,
				// Simulates `rill sudo quota set --projects 10` landing mid-flight.
				onBillingCall: func(ctx context.Context) {
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
			svc := &Service{DB: db, Biller: biller, PaymentProvider: payment.NewNoop(), Logger: zap.NewNop()}

			// Pass the snapshot read before the quota update, like the river jobs do.
			updated, err := tt.run(svc, ctx, org)
			require.NoError(t, err)

			persisted, err := db.FindOrganization(ctx, org.ID)
			require.NoError(t, err)
			for _, o := range []*database.Organization{updated, persisted} {
				require.Equal(t, 10, o.QuotaProjects, "concurrent quota update was reverted")
				require.Equal(t, tt.wantSeats, o.QuotaSeats)
				require.Equal(t, 1, o.QuotaDeployments)
				require.Equal(t, tt.wantPlanName, o.BillingPlanName)
				require.Equal(t, tt.orgName, o.BillingCustomerID)
			}
		})
	}
}

// quotaRaceBiller is a biller with a trial plan that runs onBillingCall whenever a flow calls out to it.
type quotaRaceBiller struct {
	billing.Biller
	activeSub     bool
	onBillingCall func(ctx context.Context)
}

func (b *quotaRaceBiller) GetPlanByType(ctx context.Context, planType billing.PlanType) (*billing.Plan, error) {
	return &billing.Plan{
		ID:          "trial_plan_id",
		Name:        "trial_plan",
		PlanType:    planType,
		DisplayName: "Trial",
		Quotas: billing.Quotas{
			NumProjects: new(1),
			NumSeats:    new(5),
		},
	}, nil
}

func (b *quotaRaceBiller) CreateCustomer(ctx context.Context, org *database.Organization, provider billing.PaymentProvider) (*billing.Customer, error) {
	b.onBillingCall(ctx)
	return &billing.Customer{ID: org.Name}, nil
}

func (b *quotaRaceBiller) FindCustomer(ctx context.Context, customerID string) (*billing.Customer, error) {
	return &billing.Customer{ID: customerID, PaymentProviderID: "payment"}, nil
}

func (b *quotaRaceBiller) GetActiveSubscription(ctx context.Context, customerID string) (*billing.Subscription, error) {
	b.onBillingCall(ctx)
	if !b.activeSub {
		return nil, billing.ErrNotFound
	}
	plan, err := b.GetPlanByType(ctx, billing.FreePlanType)
	if err != nil {
		return nil, err
	}
	return &billing.Subscription{ID: "sub", Plan: plan, StartDate: time.Now()}, nil
}

func (b *quotaRaceBiller) GetCustomerCreditBalance(ctx context.Context, customerID, currency string) (float64, error) {
	return billing.CreditTrialLowBalanceThreshold + 1, nil
}

func (b *quotaRaceBiller) CreateSubscription(ctx context.Context, customerID string, plan *billing.Plan) (*billing.Subscription, error) {
	b.onBillingCall(ctx)
	now := time.Now()
	return &billing.Subscription{ID: "sub", Plan: plan, StartDate: now, TrialEndDate: now.AddDate(0, 0, 30)}, nil
}
