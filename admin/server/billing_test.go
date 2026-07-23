package server_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/testadmin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	billingOpBalance = "balance"
	billingOpGrant   = "grant"
	billingOpDebit   = "debit"
	billingOpPlan    = "plan"
	billingOpCreate  = "create"
)

var errInjectedBilling = errors.New("injected billing failure")

func TestUpdateBillingSubscriptionSaga(t *testing.T) {
	// The integration boundary uses real authorization and Postgres while external billing remains deterministic and fault-injectable.
	fix := testadmin.New(t)
	_, _ = fix.NewUser(t) // The first fixture user is a bootstrap superuser; use a regular user for quota and permission assertions.
	user, owner := fix.NewUser(t)
	_, outsider := fix.NewUser(t)

	t.Run("invalid requests have no external effects", func(t *testing.T) {
		// Rejected callers and plan identifiers must not cross a mutating provider boundary.
		tests := []struct {
			name        string
			plan        string
			caller      func(*client.Client, *client.Client) *client.Client
			wantCode    codes.Code
			wantMessage string
		}{
			{name: "permission denied", plan: "paid", caller: func(_, outsider *client.Client) *client.Client { return outsider }, wantCode: codes.PermissionDenied, wantMessage: "not allowed"},
			{name: "empty plan", plan: "", caller: func(owner, _ *client.Client) *client.Client { return owner }, wantCode: codes.InvalidArgument, wantMessage: "PlanName"},
			{name: "unknown plan", plan: "missing", caller: func(owner, _ *client.Client) *client.Client { return owner }, wantCode: codes.NotFound, wantMessage: "plan not found"},
			{name: "private plan", plan: "private", caller: func(owner, _ *client.Client) *client.Client { return owner }, wantCode: codes.FailedPrecondition, wantMessage: "private plan"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Authorization and plan validation must finish before any provider mutation or payment lookup.
				org, biller := newBillingScenario(t, fix, owner)
				provider := newTestPaymentProvider(true, true)
				fix.Admin.Biller = biller
				fix.Admin.PaymentProvider = provider
				if tt.plan == "private" {
					_, err := fix.Admin.DB.UpsertBillingIssue(t.Context(), &database.UpsertBillingIssueOptions{
						OrgID: org.ID, Type: database.BillingIssueTypeNeverSubscribed, Metadata: &database.BillingIssueMetadataNeverSubscribed{}, EventTime: time.Now(),
					})
					require.NoError(t, err)
					biller.prepareUnsubscribedCustomer()
				}

				_, err := tt.caller(owner, outsider).UpdateBillingSubscription(t.Context(), &adminv1.UpdateBillingSubscriptionRequest{Org: org.Name, PlanName: tt.plan})
				require.Error(t, err)
				require.Equal(t, tt.wantCode, status.Code(err))
				require.ErrorContains(t, err, tt.wantMessage)
				require.Zero(t, biller.mutationAttempts(), "rejected requests must not mutate billing state")
				require.Zero(t, provider.findCalls(), "rejected plans must not reach payment validation")
			})
		}
	})

	t.Run("payment validation has no billing side effects", func(t *testing.T) {
		// Payment readiness is a precondition, never a partially mutating plan-change step.
		tests := []struct {
			name             string
			hasPaymentMethod bool
			hasAddress       bool
			wantMessage      string
		}{
			{name: "missing payment method", hasAddress: true, wantMessage: "no payment method found"},
			{name: "missing billing address", hasPaymentMethod: true, wantMessage: "no billing address found"},
			{name: "missing payment method and address", wantMessage: "no payment method found, no billing address found"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// A customer that cannot be billed must be rejected before credits or subscription state are touched.
				org, biller := newBillingScenario(t, fix, owner)
				provider := newTestPaymentProvider(tt.hasPaymentMethod, tt.hasAddress)
				fix.Admin.Biller = biller
				fix.Admin.PaymentProvider = provider

				_, err := owner.UpdateBillingSubscription(t.Context(), &adminv1.UpdateBillingSubscriptionRequest{Org: org.Name, PlanName: "paid"})
				require.Error(t, err)
				require.Equal(t, codes.FailedPrecondition, status.Code(err))
				require.ErrorContains(t, err, tt.wantMessage)
				require.Zero(t, biller.mutationAttempts(), "payment validation failures must not mutate the provider")
				require.Equal(t, 1, provider.findCalls())
			})
		}
	})

	t.Run("trial quota denial has no external effects", func(t *testing.T) {
		// Exhausted trial quota must stop trial creation before credits or subscriptions are provisioned.
		org, biller := newBillingScenario(t, fix, owner)
		fix.Admin.Biller = biller
		fix.Admin.PaymentProvider = newTestPaymentProvider(true, true)
		_, err := fix.Admin.DB.UpsertBillingIssue(t.Context(), &database.UpsertBillingIssueOptions{
			OrgID: org.ID, Type: database.BillingIssueTypeNeverSubscribed, Metadata: &database.BillingIssueMetadataNeverSubscribed{}, EventTime: time.Now(),
		})
		require.NoError(t, err)
		_, err = fix.Admin.DB.UpdateUser(t.Context(), user.ID, &database.UpdateUserOptions{
			DisplayName: user.DisplayName, PhotoURL: user.PhotoURL, GithubUsername: user.GithubUsername, GithubToken: user.GithubToken,
			GithubTokenExpiresOn: user.GithubTokenExpiresOn, GithubRefreshToken: user.GithubRefreshToken, QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
			QuotaTrialOrgs: 0, PreferenceTimeZone: user.PreferenceTimeZone,
		})
		require.NoError(t, err)
		updatedUser, err := fix.Admin.DB.FindUser(t.Context(), user.ID)
		require.NoError(t, err)
		require.Zero(t, updatedUser.QuotaTrialOrgs)
		require.NotNil(t, org.CreatedByUserID)

		_, err = owner.UpdateBillingSubscription(t.Context(), &adminv1.UpdateBillingSubscriptionRequest{Org: org.Name, PlanName: "trial"})
		require.Error(t, err)
		require.Equal(t, codes.FailedPrecondition, status.Code(err), err)
		require.ErrorContains(t, err, "trial orgs quota")
		require.Zero(t, biller.mutationAttempts(), "quota denial must not grant trial credits or create a subscription")
	})

	t.Run("partial failures converge on retry", func(t *testing.T) {
		// The matrix covers failures before and after every non-atomic provider mutation plus the final Postgres write.
		tests := []struct {
			name               string
			fault              *testBillingFault
			failDatabaseUpdate bool
			wantErr            string
			wantInitialCredits float64
			wantInitialUSD     float64
			wantInitialPlan    string
			wantInitialApplied map[string]int
		}{
			{name: "balance read rejected", fault: &testBillingFault{operation: billingOpBalance}, wantErr: "failed to fetch trial credit balance", wantInitialCredits: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{}},
			{name: "USD grant rejected", fault: &testBillingFault{operation: billingOpGrant}, wantErr: "failed to grant USD rollover credits", wantInitialCredits: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{}},
			{name: "USD grant response lost", fault: &testBillingFault{operation: billingOpGrant, afterApply: true}, wantErr: "failed to grant USD rollover credits", wantInitialCredits: 80, wantInitialUSD: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{billingOpGrant: 1}},
			{name: "credit debit rejected", fault: &testBillingFault{operation: billingOpDebit}, wantErr: "failed to debit trial credits", wantInitialCredits: 80, wantInitialUSD: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{billingOpGrant: 1}},
			{name: "credit debit response lost", fault: &testBillingFault{operation: billingOpDebit, afterApply: true}, wantErr: "failed to debit trial credits", wantInitialUSD: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{billingOpGrant: 1, billingOpDebit: 1}},
			{name: "plan change rejected", fault: &testBillingFault{operation: billingOpPlan}, wantErr: "injected billing failure", wantInitialUSD: 80, wantInitialPlan: "trial", wantInitialApplied: map[string]int{billingOpGrant: 1, billingOpDebit: 1}},
			{name: "plan change response lost", fault: &testBillingFault{operation: billingOpPlan, afterApply: true}, wantErr: "injected billing failure", wantInitialUSD: 80, wantInitialPlan: "paid", wantInitialApplied: map[string]int{billingOpGrant: 1, billingOpDebit: 1, billingOpPlan: 1}},
			{name: "quota database update rejected", failDatabaseUpdate: true, wantErr: "injected organization update failure", wantInitialUSD: 80, wantInitialPlan: "paid", wantInitialApplied: map[string]int{billingOpGrant: 1, billingOpDebit: 1, billingOpPlan: 1}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Each provider or database boundary must resume without duplicating money and eventually agree on the paid plan and quotas.
				org, biller := newBillingScenario(t, fix, owner)
				biller.setFault(tt.fault)
				fix.Admin.Biller = biller
				fix.Admin.PaymentProvider = newTestPaymentProvider(true, true)

				var originalDB database.DB
				if tt.failDatabaseUpdate {
					originalDB = fix.Admin.DB
					fix.Admin.DB = &failOrganizationUpdateDB{DB: originalDB, remaining: 1}
					defer func() { fix.Admin.DB = originalDB }()
				}

				_, err := owner.UpdateBillingSubscription(t.Context(), &adminv1.UpdateBillingSubscriptionRequest{Org: org.Name, PlanName: "paid"})
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				require.Equal(t, tt.wantInitialCredits, biller.balance(billing.CreditsCurrency))
				require.Equal(t, tt.wantInitialUSD, biller.balance(billing.USDCurrency))
				require.Equal(t, tt.wantInitialPlan, biller.activePlanName())
				for _, operation := range []string{billingOpGrant, billingOpDebit, billingOpPlan} {
					require.Equal(t, tt.wantInitialApplied[operation], biller.appliedCount(operation), "unexpected provider effect before retry")
				}
				stored, err := fix.Admin.DB.FindOrganization(t.Context(), org.ID)
				require.NoError(t, err)
				require.Equal(t, "trial", derefString(stored.BillingPlanName), "the DB cache must not claim success before the handler completes")

				resp, err := owner.UpdateBillingSubscription(t.Context(), &adminv1.UpdateBillingSubscriptionRequest{Org: org.Name, PlanName: "paid"})
				require.NoError(t, err)
				require.Equal(t, "paid", resp.Subscription.Plan.Name)
				require.Equal(t, float64(80), biller.balance(billing.USDCurrency))
				require.Zero(t, biller.balance(billing.CreditsCurrency))
				require.Equal(t, float64(80), biller.totalValue(), "retry must conserve the original trial-credit value")
				require.Equal(t, 1, biller.appliedCount(billingOpGrant), "the logical USD grant must apply at most once")
				require.Equal(t, 1, biller.appliedCount(billingOpDebit), "the logical trial debit must apply at most once")
				require.Equal(t, 1, biller.appliedCount(billingOpPlan), "the logical plan change must apply at most once")
				require.NotEmpty(t, biller.appliedKey(billingOpGrant), "financial retries require a stable provider idempotency key")
				require.NotEmpty(t, biller.appliedKey(billingOpDebit), "financial retries require a stable provider idempotency key")

				stored, err = fix.Admin.DB.FindOrganization(t.Context(), org.ID)
				require.NoError(t, err)
				require.Equal(t, "paid", derefString(stored.BillingPlanName))
				require.Equal(t, 7, stored.QuotaProjects)
			})
		}
	})
}

type testBillingFault struct {
	operation  string
	afterApply bool
}

type testBiller struct {
	billing.Biller

	mu          sync.Mutex
	plans       map[string]*billing.Plan
	active      *billing.Subscription
	balances    map[string]float64
	fault       *testBillingFault
	attempts    map[string]int
	applied     map[string]int
	appliedKeys map[string]string
	seenKeys    map[string]struct{}
}

func newTestBiller(org *database.Organization) *testBiller {
	trialProjects := 1
	paidProjects := 7
	trial := &billing.Plan{ID: "plan-trial", Name: "trial", DisplayName: "Trial", PlanType: billing.FreePlanType, Public: true, Quotas: billing.Quotas{NumProjects: &trialProjects}}
	paid := &billing.Plan{ID: "plan-paid", Name: "paid", DisplayName: "Paid", PlanType: billing.TeamPlanType, Public: true, Quotas: billing.Quotas{NumProjects: &paidProjects}}
	private := &billing.Plan{ID: "plan-private", Name: "private", DisplayName: "Private", PlanType: billing.FreePlanType, Public: false, Quotas: billing.Quotas{NumProjects: &paidProjects}}
	return &testBiller{
		Biller: billing.NewNoop(),
		plans: map[string]*billing.Plan{
			trial.Name:   trial,
			paid.Name:    paid,
			private.Name: private,
		},
		active: &billing.Subscription{
			ID: "subscription-" + org.ID, Customer: &billing.Customer{ID: org.BillingCustomerID}, Plan: clonePlan(trial),
			CurrentBillingCycleEndDate: time.Now().Add(30 * 24 * time.Hour),
		},
		balances:    map[string]float64{billing.CreditsCurrency: 80, billing.USDCurrency: 0},
		attempts:    make(map[string]int),
		applied:     make(map[string]int),
		appliedKeys: make(map[string]string),
		seenKeys:    make(map[string]struct{}),
	}
}

func (b *testBiller) GetPlanByName(_ context.Context, name string) (*billing.Plan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	plan := b.plans[name]
	if plan == nil {
		return nil, billing.ErrNotFound
	}
	return clonePlan(plan), nil
}

func (b *testBiller) GetPlanByType(_ context.Context, planType billing.PlanType) (*billing.Plan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, plan := range b.plans {
		if plan.PlanType == planType {
			return clonePlan(plan), nil
		}
	}
	return nil, billing.ErrNotFound
}

func (b *testBiller) GetActiveSubscription(_ context.Context, _ string) (*billing.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active == nil {
		return nil, billing.ErrNotFound
	}
	return cloneSubscription(b.active), nil
}

func (b *testBiller) GetCustomerCreditBalance(_ context.Context, _ string, currency string) (float64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts[billingOpBalance]++
	if b.consumeFault(billingOpBalance, false) {
		return 0, errInjectedBilling
	}
	return b.balances[currency], nil
}

func (b *testBiller) GrantCustomerCredits(_ context.Context, _ string, amount float64, currency, _ string, _ *time.Time, idempotencyKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts[billingOpGrant]++
	if b.alreadyApplied(billingOpGrant, idempotencyKey) {
		return nil
	}
	if b.consumeFault(billingOpGrant, false) {
		return errInjectedBilling
	}
	b.balances[currency] += amount
	b.markApplied(billingOpGrant, idempotencyKey)
	if b.consumeFault(billingOpGrant, true) {
		return errInjectedBilling
	}
	return nil
}

func (b *testBiller) DebitCustomerCredits(_ context.Context, _ string, amount float64, currency, _, idempotencyKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts[billingOpDebit]++
	if b.alreadyApplied(billingOpDebit, idempotencyKey) {
		return nil
	}
	if b.consumeFault(billingOpDebit, false) {
		return errInjectedBilling
	}
	b.balances[currency] -= amount
	b.markApplied(billingOpDebit, idempotencyKey)
	if b.consumeFault(billingOpDebit, true) {
		return errInjectedBilling
	}
	return nil
}

func (b *testBiller) ChangeSubscriptionPlan(_ context.Context, _ string, plan *billing.Plan, idempotencyKey string) (*billing.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts[billingOpPlan]++
	if b.alreadyApplied(billingOpPlan, idempotencyKey) {
		return cloneSubscription(b.active), nil
	}
	if b.consumeFault(billingOpPlan, false) {
		return nil, errInjectedBilling
	}
	b.active.Plan = clonePlan(plan)
	b.markApplied(billingOpPlan, idempotencyKey)
	if b.consumeFault(billingOpPlan, true) {
		return nil, errInjectedBilling
	}
	return cloneSubscription(b.active), nil
}

func (b *testBiller) CreateSubscription(_ context.Context, _ string, plan *billing.Plan) (*billing.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts[billingOpCreate]++
	b.applied[billingOpCreate]++
	b.active = &billing.Subscription{ID: "created-subscription", Customer: &billing.Customer{}, Plan: clonePlan(plan)}
	return cloneSubscription(b.active), nil
}

func (b *testBiller) setFault(fault *testBillingFault) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if fault == nil {
		b.fault = nil
		return
	}
	copy := *fault
	b.fault = &copy
}

func (b *testBiller) prepareUnsubscribedCustomer() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.active = nil
	b.balances[billing.CreditsCurrency] = 0
}

func (b *testBiller) consumeFault(operation string, afterApply bool) bool {
	if b.fault == nil || b.fault.operation != operation || b.fault.afterApply != afterApply {
		return false
	}
	b.fault = nil
	return true
}

func (b *testBiller) alreadyApplied(operation, key string) bool {
	if key == "" {
		return false
	}
	_, ok := b.seenKeys[operation+"\x00"+key]
	return ok
}

func (b *testBiller) markApplied(operation, key string) {
	b.applied[operation]++
	b.appliedKeys[operation] = key
	if key != "" {
		b.seenKeys[operation+"\x00"+key] = struct{}{}
	}
}

func (b *testBiller) mutationAttempts() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.attempts[billingOpGrant] + b.attempts[billingOpDebit] + b.attempts[billingOpPlan] + b.attempts[billingOpCreate]
}

func (b *testBiller) balance(currency string) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balances[currency]
}

func (b *testBiller) totalValue() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balances[billing.CreditsCurrency] + b.balances[billing.USDCurrency]
}

func (b *testBiller) activePlanName() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active.Plan.Name
}

func (b *testBiller) appliedCount(operation string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.applied[operation]
}

func (b *testBiller) appliedKey(operation string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.appliedKeys[operation]
}

type testPaymentProvider struct {
	payment.Provider

	mu       sync.Mutex
	customer payment.Customer
	find     int
}

func newTestPaymentProvider(hasPaymentMethod, hasAddress bool) *testPaymentProvider {
	return &testPaymentProvider{
		Provider: payment.NewNoop(),
		customer: payment.Customer{ID: "payment-customer", HasPaymentMethod: hasPaymentMethod, HasBillableAddress: hasAddress},
	}
}

func (p *testPaymentProvider) FindCustomer(_ context.Context, _ string) (*payment.Customer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.find++
	copy := p.customer
	return &copy, nil
}

func (p *testPaymentProvider) findCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.find
}

type failOrganizationUpdateDB struct {
	database.DB

	mu        sync.Mutex
	remaining int
}

func (db *failOrganizationUpdateDB) UpdateOrganization(ctx context.Context, id string, opts *database.UpdateOrganizationOptions) (*database.Organization, error) {
	db.mu.Lock()
	if db.remaining > 0 {
		db.remaining--
		db.mu.Unlock()
		return nil, errors.New("injected organization update failure")
	}
	db.mu.Unlock()
	return db.DB.UpdateOrganization(ctx, id, opts)
}

func newBillingScenario(t *testing.T, fix *testadmin.Fixture, owner *client.Client) (*database.Organization, *testBiller) {
	t.Helper()
	resp, err := owner.CreateOrganization(t.Context(), &adminv1.CreateOrganizationRequest{Name: randomName()})
	require.NoError(t, err)
	org, err := fix.Admin.DB.FindOrganizationByName(t.Context(), resp.Organization.Name)
	require.NoError(t, err)
	trialName := "trial"
	trialDisplayName := "Trial"
	org, err = fix.Admin.DB.UpdateOrganization(t.Context(), org.ID, &database.UpdateOrganizationOptions{
		Name: org.Name, DisplayName: org.DisplayName, Description: org.Description, LogoAssetID: org.LogoAssetID, LogoDarkAssetID: org.LogoDarkAssetID,
		FaviconAssetID: org.FaviconAssetID, ThumbnailAssetID: org.ThumbnailAssetID, CustomDomain: org.CustomDomain, DefaultProjectRoleID: org.DefaultProjectRoleID,
		QuotaProjects: org.QuotaProjects, QuotaDeployments: org.QuotaDeployments, QuotaSlotsTotal: org.QuotaSlotsTotal, QuotaSlotsPerDeployment: org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites: org.QuotaOutstandingInvites, QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment, QuotaSeats: org.QuotaSeats,
		BillingCustomerID: "billing-" + org.ID, PaymentCustomerID: "payment-" + org.ID, BillingEmail: org.BillingEmail,
		BillingPlanName: &trialName, BillingPlanDisplayName: &trialDisplayName, CreatedByUserID: org.CreatedByUserID,
	})
	require.NoError(t, err)
	return org, newTestBiller(org)
}

func clonePlan(plan *billing.Plan) *billing.Plan {
	if plan == nil {
		return nil
	}
	copy := *plan
	return &copy
}

func cloneSubscription(sub *billing.Subscription) *billing.Subscription {
	if sub == nil {
		return nil
	}
	copy := *sub
	copy.Plan = clonePlan(sub.Plan)
	if sub.Customer != nil {
		customer := *sub.Customer
		copy.Customer = &customer
	}
	return &copy
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

var _ billing.Biller = (*testBiller)(nil)
var _ payment.Provider = (*testPaymentProvider)(nil)
var _ database.DB = (*failOrganizationUpdateDB)(nil)
