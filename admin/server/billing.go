package server

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetBillingSubscription(ctx context.Context, req *adminv1.GetBillingSubscriptionRequest) (*adminv1.GetBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org subscriptions")
	}

	if org.BillingCustomerID == "" {
		return &adminv1.GetBillingSubscriptionResponse{Organization: organizationToDTO(org)}, nil
	}

	subs, err := s.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(subs) == 0 {
		return &adminv1.GetBillingSubscriptionResponse{Organization: organizationToDTO(org)}, nil
	}

	if len(subs) > 1 {
		s.logger.Warn("multiple subscriptions found for the organization", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	return &adminv1.GetBillingSubscriptionResponse{
		Organization:     organizationToDTO(org),
		Subscription:     subscriptionToDTO(subs[0]),
		BillingPortalUrl: subs[0].Customer.PortalURL,
	}, nil
}

func (s *Server) UpdateBillingSubscription(ctx context.Context, req *adminv1.UpdateBillingSubscriptionRequest) (*adminv1.UpdateBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	if req.PlanName != "" {
		observability.AddRequestAttributes(ctx, attribute.String("args.plan_name", req.PlanName))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org billing plan")
	}

	if req.PlanName == "" {
		return nil, status.Error(codes.InvalidArgument, "plan name must be provided")
	}

	plan, err := s.admin.Biller.GetPlanByName(ctx, req.PlanName)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "plan not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	org, subs, err := s.admin.RepairOrgBilling(ctx, org)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, sub := range subs {
		if sub.Plan.ID == plan.ID {
			return nil, status.Errorf(codes.FailedPrecondition, "organization already subscribed to the plan %s", plan.Name)
		}
	}

	// plan change needed
	// check for a payment method and a valid billing address
	var validationErrs []string
	pc, err := s.admin.PaymentProvider.FindCustomer(ctx, org.PaymentCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !pc.HasPaymentMethod {
		validationErrs = append(validationErrs, "no payment method found")
	}

	bc, err := s.admin.Biller.FindCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !bc.HasBillableAddress {
		validationErrs = append(validationErrs, "no billing address found, click on update information to add billing address")
	}

	be, err := s.admin.DB.FindBillingIssueByType(ctx, org.ID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if be != nil {
		validationErrs = append(validationErrs, "a previous payment is due, please pay the outstanding amount")
	}

	if len(validationErrs) > 0 && !claims.Superuser(ctx) {
		return nil, status.Errorf(codes.FailedPrecondition, "please fix following by visiting billing portal: %s", strings.Join(validationErrs, ", "))
	}

	if planDowngrade(plan, org) {
		if claims.Superuser(ctx) {
			s.logger.Named("billing").Warn("plan downgraded", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("current_plan_id", subs[0].Plan.ID), zap.String("current_plan_name", subs[0].Plan.Name), zap.String("new_plan_id", plan.ID), zap.String("new_plan_name", plan.Name))
		} else {
			return nil, status.Errorf(codes.FailedPrecondition, "plan downgrade not supported")
		}
	}

	// TODO move below to background job
	if len(subs) == 1 {
		// schedule plan change
		_, err = s.admin.Biller.ChangeSubscriptionPlan(ctx, subs[0].ID, plan, billing.SubscriptionChangeOptionImmediate)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		// multiple subscriptions, cancel them first immediately and assign new plan should not happen unless externally assigned multiple subscriptions to the same org in the billing system.
		// RepairOrgBilling does not fix multiple subscription issue, we are not sure which subscription to cancel and which to keep. However, in case of plan change we can safely cancel all older subscriptions and create a new one with new plan.
		for _, sub := range subs {
			err = s.admin.Biller.CancelSubscription(ctx, sub.ID, billing.SubscriptionCancellationOptionImmediate)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		// create new subscription
		_, err = s.admin.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	s.logger.Info("plan changed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("current_plan_id", subs[0].Plan.ID), zap.String("current_plan_name", subs[0].Plan.Name), zap.String("new_plan_id", plan.ID), zap.String("new_plan_name", plan.Name))

	// schedule plan change by API job
	_, err = s.admin.Jobs.PlanChangeByAPI(ctx, org.ID, subs[0].ID, plan.ID, subs[0].CurrentBillingCycleStartDate)
	if err != nil {
		return nil, err
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       valOrDefault(plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    valOrDefault(plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     valOrDefault(plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             valOrDefault(plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             valOrDefault(plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subs, err = s.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var subscriptions []*adminv1.Subscription
	for _, sub := range subs {
		subscriptions = append(subscriptions, subscriptionToDTO(sub))
	}

	return &adminv1.UpdateBillingSubscriptionResponse{
		Organization:  organizationToDTO(org),
		Subscriptions: subscriptions,
	}, nil
}

// CancelBillingSubscription cancels the billing subscription for the organization and puts them on default plan
func (s *Server) CancelBillingSubscription(ctx context.Context, req *adminv1.CancelBillingSubscriptionRequest) (*adminv1.CancelBillingSubscriptionResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to cancel org subscription")
	}

	subs, err := s.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(subs) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no subscription found for the organization")
	}

	plan, err := s.admin.Biller.GetDefaultPlan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subEndDate := subs[0].CurrentBillingCycleEndDate
	if plan.ID != subs[0].Plan.ID {
		// schedule plan change to default plan at end of the current subscription term
		if len(subs) == 1 {
			_, err = s.admin.Biller.ChangeSubscriptionPlan(ctx, subs[0].ID, plan, billing.SubscriptionChangeOptionEndOfSubscriptionTerm)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			// multiple subscriptions, cancel them first immediately and assign default plan to start at the end of the current subscription term
			latestEndDate := time.Time{}
			for _, sub := range subs {
				if sub.CurrentBillingCycleEndDate.After(latestEndDate) {
					latestEndDate = sub.CurrentBillingCycleEndDate
				}
				err = s.admin.Biller.CancelSubscription(ctx, sub.ID, billing.SubscriptionCancellationOptionEndOfSubscriptionTerm)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
			}

			// create new subscription with default plan to start at the end of the current subscription term
			_, err = s.admin.Biller.CreateSubscriptionInFuture(ctx, org.BillingCustomerID, plan, latestEndDate.AddDate(0, 0, 1))
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			subEndDate = latestEndDate
		}
	}

	// raise a billing issue of the subscription cancellation
	_, err = s.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeSubscriptionCancelled,
		Metadata: database.BillingIssueMetadataSubscriptionCancelled{
			EndDate: subEndDate,
		},
		EventTime: time.Now(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// clean up any trial related billing errors and warnings if present
	err = s.admin.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CancelBillingSubscriptionResponse{}, nil
}

func (s *Server) GetPaymentsPortalURL(ctx context.Context, req *adminv1.GetPaymentsPortalURLRequest) (*adminv1.GetPaymentsPortalURLResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	observability.AddRequestAttributes(ctx, attribute.String("args.return_url", req.ReturnUrl))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage org billing")
	}

	if org.PaymentCustomerID == "" {
		_, _, err = s.admin.RepairOrgBilling(ctx, org)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	url, err := s.admin.PaymentProvider.GetBillingPortalURL(ctx, org.PaymentCustomerID, req.ReturnUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.GetPaymentsPortalURLResponse{Url: url}, nil
}

// SudoUpdateOrganizationBillingCustomer updates the billing customer id for an organization. May be useful if customer is initialized manually in billing system
func (s *Server) SudoUpdateOrganizationBillingCustomer(ctx context.Context, req *adminv1.SudoUpdateOrganizationBillingCustomerRequest) (*adminv1.SudoUpdateOrganizationBillingCustomerResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.billing_customer_id", req.BillingCustomerId),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage billing customer")
	}

	if req.BillingCustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "billing customer id is required")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	opts := &database.UpdateOrganizationOptions{
		Name:                                req.Organization,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   req.BillingCustomerId,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, opts)
	if err != nil {
		return nil, err
	}

	// get subscriptions if present
	subs, err := s.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var subscriptions []*adminv1.Subscription
	for _, sub := range subs {
		subscriptions = append(subscriptions, subscriptionToDTO(sub))
	}

	return &adminv1.SudoUpdateOrganizationBillingCustomerResponse{
		Organization:  organizationToDTO(org),
		Subscriptions: subscriptions,
	}, nil
}

func (s *Server) ListPublicBillingPlans(ctx context.Context, req *adminv1.ListPublicBillingPlansRequest) (*adminv1.ListPublicBillingPlansResponse, error) {
	observability.AddRequestAttributes(ctx)

	// no permissions required to list public billing plans
	plans, err := s.admin.Biller.GetPublicPlans(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dtos []*adminv1.BillingPlan
	for _, plan := range plans {
		dtos = append(dtos, billingPlanToDTO(plan))
	}

	return &adminv1.ListPublicBillingPlansResponse{
		Plans: dtos,
	}, nil
}

func (s *Server) ListOrganizationBillingIssues(ctx context.Context, req *adminv1.ListOrganizationBillingIssuesRequest) (*adminv1.ListOrganizationBillingIssuesResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org billing errors")
	}

	issues, err := s.admin.DB.FindBillingIssues(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dtos []*adminv1.BillingIssue
	for _, i := range issues {
		dtos = append(dtos, &adminv1.BillingIssue{
			Organization: org.Name,
			Type:         billingIssueTypeToDTO(i.Type),
			Level:        billingIssueLevelToDTO(i.Level),
			Metadata:     billingIssueMetadataToDTO(i.Type, i.Metadata),
			EventTime:    timestamppb.New(i.EventTime),
			CreatedOn:    timestamppb.New(i.CreatedOn),
		})
	}

	return &adminv1.ListOrganizationBillingIssuesResponse{
		Issues: dtos,
	}, nil
}

func (s *Server) SudoDeleteOrganizationBillingIssue(ctx context.Context, req *adminv1.SudoDeleteOrganizationBillingIssueRequest) (*adminv1.SudoDeleteOrganizationBillingIssueResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization), attribute.String("args.type", req.Type.String()))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can delete billing errors")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	t, err := dtoBillingIssueTypeToDB(req.Type)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteBillingIssueByType(ctx, org.ID, t)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SudoDeleteOrganizationBillingIssueResponse{}, nil
}

func subscriptionToDTO(sub *billing.Subscription) *adminv1.Subscription {
	return &adminv1.Subscription{
		Id:                           sub.ID,
		PlanId:                       sub.Plan.ID,
		PlanName:                     sub.Plan.Name,
		PlanDisplayName:              sub.Plan.DisplayName,
		StartDate:                    valOrNullTime(sub.StartDate),
		EndDate:                      valOrNullTime(sub.EndDate),
		CurrentBillingCycleStartDate: valOrNullTime(sub.CurrentBillingCycleStartDate),
		CurrentBillingCycleEndDate:   valOrNullTime(sub.CurrentBillingCycleEndDate),
		TrialEndDate:                 valOrNullTime(sub.TrialEndDate),
	}
}

func billingPlanToDTO(plan *billing.Plan) *adminv1.BillingPlan {
	return &adminv1.BillingPlan{
		Id:              plan.ID,
		Name:            plan.Name,
		DisplayName:     plan.DisplayName,
		Description:     plan.Description,
		TrialPeriodDays: uint32(plan.TrialPeriodDays),
		Default:         plan.Default,
		Quotas: &adminv1.Quotas{
			Projects:                       valOrEmptyString(plan.Quotas.NumProjects),
			Deployments:                    valOrEmptyString(plan.Quotas.NumDeployments),
			SlotsTotal:                     valOrEmptyString(plan.Quotas.NumSlotsTotal),
			SlotsPerDeployment:             valOrEmptyString(plan.Quotas.NumSlotsPerDeployment),
			OutstandingInvites:             valOrEmptyString(plan.Quotas.NumOutstandingInvites),
			StorageLimitBytesPerDeployment: val64OrEmptyString(plan.Quotas.StorageLimitBytesPerDeployment),
		},
	}
}

func valOrNullTime(v time.Time) *timestamppb.Timestamp {
	if v.IsZero() {
		return nil
	}
	return timestamppb.New(v)
}

func billingIssueTypeToDTO(t database.BillingIssueType) adminv1.BillingIssueType {
	switch t {
	case database.BillingIssueTypeOnTrial:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL
	case database.BillingIssueTypeTrialEnded:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED
	case database.BillingIssueTypeNoPaymentMethod:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD
	case database.BillingIssueTypeNoBillableAddress:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS
	case database.BillingIssueTypePaymentFailed:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED
	case database.BillingIssueTypeSubscriptionCancelled:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED
	default:
		return adminv1.BillingIssueType_BILLING_ISSUE_TYPE_UNSPECIFIED
	}
}

func billingIssueLevelToDTO(l database.BillingIssueLevel) adminv1.BillingIssueLevel {
	switch l {
	case database.BillingIssueLevelError:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_ERROR
	case database.BillingIssueLevelWarning:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_WARNING
	default:
		return adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_UNSPECIFIED
	}
}

func dtoBillingIssueTypeToDB(t adminv1.BillingIssueType) (database.BillingIssueType, error) {
	switch t {
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL:
		return database.BillingIssueTypeOnTrial, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED:
		return database.BillingIssueTypeTrialEnded, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_PAYMENT_METHOD:
		return database.BillingIssueTypeNoPaymentMethod, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_NO_BILLABLE_ADDRESS:
		return database.BillingIssueTypeNoBillableAddress, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED:
		return database.BillingIssueTypePaymentFailed, nil
	case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED:
		return database.BillingIssueTypeSubscriptionCancelled, nil
	default:
		return database.BillingIssueTypeUnspecified, status.Error(codes.InvalidArgument, "invalid billing error type")
	}
}

func billingIssueMetadataToDTO(t database.BillingIssueType, m database.BillingIssueMetadata) *adminv1.BillingIssueMetadata {
	switch t {
	case database.BillingIssueTypeOnTrial:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_OnTrial{
				OnTrial: &adminv1.BillingIssueMetadataOnTrial{
					EndDate: timestamppb.New(m.(*database.BillingIssueMetadataOnTrial).EndDate),
				},
			},
		}
	case database.BillingIssueTypeTrialEnded:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_TrialEnded{
				TrialEnded: &adminv1.BillingIssueMetadataTrialEnded{
					GracePeriodEndDate: timestamppb.New(m.(*database.BillingIssueMetadataTrialEnded).GracePeriodEndDate),
				},
			},
		}
	case database.BillingIssueTypeNoPaymentMethod:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_NoPaymentMethod{
				NoPaymentMethod: &adminv1.BillingIssueMetadataNoPaymentMethod{},
			},
		}
	case database.BillingIssueTypeNoBillableAddress:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_NoBillableAddress{
				NoBillableAddress: &adminv1.BillingIssueMetadataNoBillableAddress{},
			},
		}
	case database.BillingIssueTypePaymentFailed:
		paymentFailed := m.(*database.BillingIssueMetadataPaymentFailed)
		invoices := make([]*adminv1.BillingIssueMetadataPaymentFailedMeta, 0)
		for k := range paymentFailed.Invoices {
			invoices = append(invoices, &adminv1.BillingIssueMetadataPaymentFailedMeta{
				InvoiceId:     paymentFailed.Invoices[k].ID,
				InvoiceNumber: paymentFailed.Invoices[k].Number,
				InvoiceUrl:    paymentFailed.Invoices[k].URL,
				AmountDue:     paymentFailed.Invoices[k].Amount,
				Currency:      paymentFailed.Invoices[k].Currency,
				DueDate:       timestamppb.New(paymentFailed.Invoices[k].DueDate),
			})
		}
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_PaymentFailed{
				PaymentFailed: &adminv1.BillingIssueMetadataPaymentFailed{
					Invoices: invoices,
				},
			},
		}
	case database.BillingIssueTypeSubscriptionCancelled:
		return &adminv1.BillingIssueMetadata{
			Metadata: &adminv1.BillingIssueMetadata_SubscriptionCancelled{
				SubscriptionCancelled: &adminv1.BillingIssueMetadataSubscriptionCancelled{},
			},
		}
	default:
		return &adminv1.BillingIssueMetadata{}
	}
}

func planDowngrade(newPan *billing.Plan, org *database.Organization) bool {
	// nil or negative values are considered as unlimited
	if comparableInt(newPan.Quotas.NumProjects) < comparableInt(&org.QuotaProjects) {
		return true
	}
	if comparableInt(newPan.Quotas.NumDeployments) < comparableInt(&org.QuotaDeployments) {
		return true
	}
	if comparableInt(newPan.Quotas.NumSlotsTotal) < comparableInt(&org.QuotaSlotsTotal) {
		return true
	}
	if comparableInt(newPan.Quotas.NumSlotsPerDeployment) < comparableInt(&org.QuotaSlotsPerDeployment) {
		return true
	}
	if comparableInt(newPan.Quotas.NumOutstandingInvites) < comparableInt(&org.QuotaOutstandingInvites) {
		return true
	}
	if comparableInt64(newPan.Quotas.StorageLimitBytesPerDeployment) < comparableInt64(&org.QuotaStorageLimitBytesPerDeployment) {
		return true
	}
	return false
}

func comparableInt(v *int) int {
	if v == nil || *v < 0 {
		return math.MaxInt
	}
	return *v
}

func comparableInt64(v *int64) int64 {
	if v == nil || *v < 0 {
		return math.MaxInt64
	}
	return *v
}
