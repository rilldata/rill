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
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.OrgName))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrgName)
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
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.OrgName))
	if req.PlanName != "" {
		observability.AddRequestAttributes(ctx, attribute.String("args.plan_name", req.PlanName))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrgName)
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

	if len(validationErrs) > 0 && !claims.Superuser(ctx) {
		return nil, status.Errorf(codes.FailedPrecondition, "please fix following by visiting billing portal: %s", strings.Join(validationErrs, ", "))
	}

	if planDowngrade(plan, org) {
		if claims.Superuser(ctx) {
			s.logger.Warn("plan downgraded", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("current_plan_id", subs[0].Plan.ID), zap.String("current_plan_name", subs[0].Plan.Name), zap.String("new_plan_id", plan.ID), zap.String("new_plan_name", plan.Name))
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
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.OrgName))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrgName)
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

	// schedule subscription cancellation job at end of the current subscription term + 1 hour
	j, err := s.admin.Jobs.SubscriptionCancellation(ctx, org.ID, subs[0].ID, plan.ID, subEndDate)
	if err != nil {
		return nil, err
	}

	// raise a billing error of the subscription cancellation
	_, err = s.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID: org.ID,
		Type:  database.BillingErrorTypeSubscriptionCancelled,
		Metadata: database.BillingErrorMetadataSubscriptionCancelled{
			EndDate:     subEndDate,
			SubEndJobID: j.ID,
		},
		EventTime: time.Now(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// clean up any trial related billing errors and warnings if present
	err = s.admin.CleanupTrialBillingErrorsAndWarnings(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CancelBillingSubscriptionResponse{}, nil
}

func (s *Server) GetPaymentsPortalURL(ctx context.Context, req *adminv1.GetPaymentsPortalURLRequest) (*adminv1.GetPaymentsPortalURLResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.OrgName))
	observability.AddRequestAttributes(ctx, attribute.String("args.return_url", req.ReturnUrl))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrgName)
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
		attribute.String("args.org", req.OrgName),
		attribute.String("args.billing_customer_id", req.BillingCustomerId),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage billing customer")
	}

	if req.BillingCustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "billing customer id is required")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrgName)
	if err != nil {
		return nil, err
	}

	opts := &database.UpdateOrganizationOptions{
		Name:                                req.OrgName,
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

func (s *Server) ListOrganizationBillingErrors(ctx context.Context, req *adminv1.ListOrganizationBillingErrorsRequest) (*adminv1.ListOrganizationBillingErrorsResponse, error) {
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

	errs, err := s.admin.DB.FindBillingErrors(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dtos []*adminv1.BillingError
	for _, e := range errs {
		dtos = append(dtos, &adminv1.BillingError{
			Organization: org.Name,
			Type:         billingErrorTypeToDTO(e.Type),
			Metadata:     billingErrorMetadataToDTO(e.Type, e.Metadata),
			EventTime:    timestamppb.New(e.EventTime),
			CreatedOn:    timestamppb.New(e.CreatedOn),
		})
	}

	return &adminv1.ListOrganizationBillingErrorsResponse{
		Errors: dtos,
	}, nil
}

func (s *Server) ListOrganizationBillingWarnings(ctx context.Context, req *adminv1.ListOrganizationBillingWarningsRequest) (*adminv1.ListOrganizationBillingWarningsResponse, error) {
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
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org billing warnings")
	}

	warnings, err := s.admin.DB.FindBillingWarnings(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dtos []*adminv1.BillingWarning
	for _, w := range warnings {
		dtos = append(dtos, &adminv1.BillingWarning{
			Organization: org.Name,
			Type:         billingWarningTypeToDTO(w.Type),
			Metadata:     billingWarningMetadataToDTO(w.Type, w.Metadata),
			EventTime:    timestamppb.New(w.EventTime),
			CreatedOn:    timestamppb.New(w.CreatedOn),
		})
	}

	return &adminv1.ListOrganizationBillingWarningsResponse{
		Warnings: dtos,
	}, nil
}

func (s *Server) SudoDeleteOrganizationBillingError(ctx context.Context, req *adminv1.SudoDeleteOrganizationBillingErrorRequest) (*adminv1.SudoDeleteOrganizationBillingErrorResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization), attribute.String("args.type", req.Type.String()))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can delete billing errors")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	t, err := dtoBillingErrorTypeToDB(req.Type)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteBillingErrorByType(ctx, org.ID, t)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SudoDeleteOrganizationBillingErrorResponse{}, nil
}

func (s *Server) SudoDeleteOrganizationBillingWarning(ctx context.Context, req *adminv1.SudoDeleteOrganizationBillingWarningRequest) (*adminv1.SudoDeleteOrganizationBillingWarningResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization), attribute.String("args.type", req.Type.String()))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can delete billing warnings")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	t, err := dtoBillingWarningTypeToDB(req.Type)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteBillingWarningByType(ctx, org.ID, t)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SudoDeleteOrganizationBillingWarningResponse{}, nil
}

func subscriptionToDTO(sub *billing.Subscription) *adminv1.Subscription {
	return &adminv1.Subscription{
		Id:                           sub.ID,
		PlanId:                       sub.Plan.ID,
		PlanName:                     sub.Plan.Name,
		PlanDisplayName:              sub.Plan.DisplayName,
		StartDate:                    timestamppb.New(sub.StartDate),
		EndDate:                      timestamppb.New(sub.EndDate),
		CurrentBillingCycleStartDate: timestamppb.New(sub.CurrentBillingCycleStartDate),
		CurrentBillingCycleEndDate:   timestamppb.New(sub.CurrentBillingCycleEndDate),
		TrialEndDate:                 timestamppb.New(sub.TrialEndDate),
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

func billingErrorTypeToDTO(t database.BillingErrorType) adminv1.BillingErrorType {
	switch t {
	case database.BillingErrorTypeUnspecified:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_UNSPECIFIED
	case database.BillingErrorTypeNoPaymentMethod:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_NO_PAYMENT_METHOD
	case database.BillingErrorTypeNoBillableAddress:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_NO_BILLABLE_ADDRESS
	case database.BillingErrorTypeTrialEnded:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_TRIAL_ENDED
	case database.BillingErrorTypeInvoicePaymentFailed:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_INVOICE_PAYMENT_FAILED
	case database.BillingErrorTypeSubscriptionCancelled:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_SUBSCRIPTION_CANCELLED
	default:
		return adminv1.BillingErrorType_BILLING_ERROR_TYPE_UNSPECIFIED
	}
}

func dtoBillingErrorTypeToDB(t adminv1.BillingErrorType) (database.BillingErrorType, error) {
	switch t {
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_UNSPECIFIED:
		return database.BillingErrorTypeUnspecified, nil
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_NO_PAYMENT_METHOD:
		return database.BillingErrorTypeNoPaymentMethod, nil
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_NO_BILLABLE_ADDRESS:
		return database.BillingErrorTypeNoBillableAddress, nil
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_TRIAL_ENDED:
		return database.BillingErrorTypeTrialEnded, nil
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_INVOICE_PAYMENT_FAILED:
		return database.BillingErrorTypeInvoicePaymentFailed, nil
	case adminv1.BillingErrorType_BILLING_ERROR_TYPE_SUBSCRIPTION_CANCELLED:
		return database.BillingErrorTypeSubscriptionCancelled, nil
	default:
		return database.BillingErrorTypeUnspecified, status.Error(codes.InvalidArgument, "invalid billing error type")
	}
}

func billingWarningTypeToDTO(t database.BillingWarningType) adminv1.BillingWarningType {
	switch t {
	case database.BillingWarningTypeUnspecified:
		return adminv1.BillingWarningType_BILLING_WARNING_TYPE_UNSPECIFIED
	case database.BillingWarningTypeOnTrial:
		return adminv1.BillingWarningType_BILLING_WARNING_TYPE_ON_TRIAL
	default:
		return adminv1.BillingWarningType_BILLING_WARNING_TYPE_UNSPECIFIED
	}
}

func dtoBillingWarningTypeToDB(t adminv1.BillingWarningType) (database.BillingWarningType, error) {
	switch t {
	case adminv1.BillingWarningType_BILLING_WARNING_TYPE_UNSPECIFIED:
		return database.BillingWarningTypeUnspecified, nil
	case adminv1.BillingWarningType_BILLING_WARNING_TYPE_ON_TRIAL:
		return database.BillingWarningTypeOnTrial, nil
	default:
		return database.BillingWarningTypeUnspecified, status.Error(codes.InvalidArgument, "invalid billing warning type")
	}
}

func billingErrorMetadataToDTO(t database.BillingErrorType, m database.BillingErrorMetadata) *adminv1.BillingErrorMetadata {
	switch t {
	case database.BillingErrorTypeUnspecified:
		return &adminv1.BillingErrorMetadata{}
	case database.BillingErrorTypeNoPaymentMethod:
		return &adminv1.BillingErrorMetadata{
			Metadata: &adminv1.BillingErrorMetadata_NoPaymentMethod{
				NoPaymentMethod: &adminv1.BillingErrorMetadataNoPaymentMethod{},
			},
		}
	case database.BillingErrorTypeNoBillableAddress:
		return &adminv1.BillingErrorMetadata{
			Metadata: &adminv1.BillingErrorMetadata_NoBillableAddress{
				NoBillableAddress: &adminv1.BillingErrorMetadataNoBillableAddress{},
			},
		}
	case database.BillingErrorTypeTrialEnded:
		return &adminv1.BillingErrorMetadata{
			Metadata: &adminv1.BillingErrorMetadata_TrialEnded{
				TrialEnded: &adminv1.BillingErrorMetadataTrialEnded{
					GracePeriodEndDate: timestamppb.New(m.(*database.BillingErrorMetadataTrialEnded).GracePeriodEndDate),
				},
			},
		}
	case database.BillingErrorTypeInvoicePaymentFailed:
		invoicePaymentFailed := m.(*database.BillingErrorMetadataInvoicePaymentFailed)
		invoices := make([]*adminv1.InvoicePaymentFailedMeta, 0)
		for k := range invoicePaymentFailed.Invoices {
			invoices = append(invoices, &adminv1.InvoicePaymentFailedMeta{
				InvoiceId:     invoicePaymentFailed.Invoices[k].ID,
				InvoiceNumber: invoicePaymentFailed.Invoices[k].Number,
				InvoiceUrl:    invoicePaymentFailed.Invoices[k].URL,
				AmountDue:     invoicePaymentFailed.Invoices[k].Amount,
				Currency:      invoicePaymentFailed.Invoices[k].Currency,
				DueDate:       timestamppb.New(invoicePaymentFailed.Invoices[k].DueDate),
			})
		}
		return &adminv1.BillingErrorMetadata{
			Metadata: &adminv1.BillingErrorMetadata_InvoicePaymentFailed{
				InvoicePaymentFailed: &adminv1.BillingErrorMetadataInvoicePaymentFailed{
					Invoices: invoices,
				},
			},
		}
	case database.BillingErrorTypeSubscriptionCancelled:
		return &adminv1.BillingErrorMetadata{
			Metadata: &adminv1.BillingErrorMetadata_SubscriptionCancelled{
				SubscriptionCancelled: &adminv1.BillingErrorMetadataSubscriptionCancelled{},
			},
		}
	default:
		return &adminv1.BillingErrorMetadata{}
	}
}

func billingWarningMetadataToDTO(t database.BillingWarningType, m database.BillingWarningMetadata) *adminv1.BillingWarningMetadata {
	switch t {
	case database.BillingWarningTypeUnspecified:
		return &adminv1.BillingWarningMetadata{}
	case database.BillingWarningTypeOnTrial:
		return &adminv1.BillingWarningMetadata{
			Metadata: &adminv1.BillingWarningMetadata_OnTrial{
				OnTrial: &adminv1.BillingWarningMetadataOnTrial{
					EndDate: timestamppb.New(m.(*database.BillingWarningMetadataOnTrial).EndDate),
				},
			},
		}
	default:
		return &adminv1.BillingWarningMetadata{}
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
