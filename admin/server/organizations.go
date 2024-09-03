package server

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/publicemail"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListOrganizations(ctx context.Context, req *adminv1.ListOrganizationsRequest) (*adminv1.ListOrganizationsResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	pageSize := validPageSize(req.PageSize)

	orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, claims.OwnerID(), token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(orgs) >= pageSize {
		nextToken = marshalPageToken(orgs[len(orgs)-1].Name)
	}

	pbs := make([]*adminv1.Organization, len(orgs))
	for i, org := range orgs {
		pbs[i] = organizationToDTO(org)
	}

	return &adminv1.ListOrganizationsResponse{Organizations: pbs, NextPageToken: nextToken}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *adminv1.GetOrganizationRequest) (*adminv1.GetOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrg && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org")
	}

	return &adminv1.GetOrganizationResponse{
		Organization: organizationToDTO(org),
		Permissions:  claims.OrganizationPermissions(ctx, org.ID),
	}, nil
}

func (s *Server) GetOrganizationNameForDomain(ctx context.Context, req *adminv1.GetOrganizationNameForDomainRequest) (*adminv1.GetOrganizationNameForDomainResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.domain", req.Domain))

	org, err := s.admin.DB.FindOrganizationByCustomDomain(ctx, req.Domain)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// NOTE: Not checking auth on purpose. This needs to be a public endpoint.

	return &adminv1.GetOrganizationNameForDomainResponse{
		Name: org.Name,
	}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Name),
		attribute.String("args.description", req.Description),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// check single user org limit for this user
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	count, err := s.admin.DB.CountSingleuserOrganizationsForMemberUser(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user.QuotaSingleuserOrgs >= 0 && count >= user.QuotaSingleuserOrgs {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: you can only create %d single-user orgs", user.QuotaSingleuserOrgs)
	}

	org, err := s.admin.CreateOrganizationForUser(ctx, user.ID, user.Email, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: organizationToDTO(org),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org")
	}

	err = s.admin.DB.DeleteOrganization(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// cancel subscription
	if org.BillingCustomerID != "" {
		err = s.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
		if err != nil {
			s.logger.Error("failed to cancel subscriptions", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}
		s.logger.Warn("canceled subscriptions", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	return &adminv1.DeleteOrganizationResponse{}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *adminv1.UpdateOrganizationRequest) (*adminv1.UpdateOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))
	if req.Description != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.description", *req.Description))
	}
	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}
	if req.BillingEmail != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.billing_email", *req.BillingEmail))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	nameChanged := req.NewName != nil && *req.NewName != org.Name
	emailChanged := req.BillingEmail != nil && *req.BillingEmail != org.BillingEmail
	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                valOrDefault(req.NewName, org.Name),
		DisplayName:                         valOrDefault(req.DisplayName, org.DisplayName),
		Description:                         valOrDefault(req.Description, org.Description),
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        valOrDefault(req.BillingEmail, org.BillingEmail),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if nameChanged {
		err := s.admin.UpdateOrgDeploymentAnnotations(ctx, org)
		if err != nil {
			return nil, err
		}
	}

	if emailChanged {
		err = s.admin.Biller.UpdateCustomerEmail(ctx, org.BillingCustomerID, org.BillingEmail)
		if err != nil {
			s.logger.Error("failed to update billing email", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.PaymentProvider.UpdateCustomerEmail(ctx, org.PaymentCustomerID, org.BillingEmail)
		if err != nil {
			s.logger.Error("failed to update billing email", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &adminv1.UpdateOrganizationResponse{
		Organization: organizationToDTO(org),
	}, nil
}

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
	_, err = riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.HandlePlanChangeByAPIArgs{
		OrgID:  org.ID,
		SubID:  subs[0].ID,
		PlanID: plan.ID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
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

	// schedule subscription cancellation job at end of the current subscription term
	j, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.HandleSubscriptionCancellationArgs{
		OrgID:  org.ID,
		SubID:  subs[0].ID,
		PlanID: plan.ID,
	}, &river.InsertOpts{
		ScheduledAt: subEndDate.AddDate(0, 0, 1).Add(1 * time.Hour), // 1 hour after the end of the current subscription term
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	// raise a billing error of the subscription cancellation
	_, err = s.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID: org.ID,
		Type:  database.BillingErrorTypeSubscriptionCancelled,
		Metadata: database.BillingErrorMetadataSubscriptionCancelled{
			EndDate:     subEndDate,
			SubEndJobID: j.Job.ID,
		},
		EventTime: time.Now(),
	})
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

func (s *Server) ListOrganizationMemberUsers(ctx context.Context, req *adminv1.ListOrganizationMemberUsersRequest) (*adminv1.ListOrganizationMemberUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) && !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	members, err := s.admin.DB.FindOrganizationMemberUsers(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.MemberUser, len(members))
	for i, user := range members {
		dtos[i] = memberUserToPB(user)
	}

	return &adminv1.ListOrganizationMemberUsersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListOrganizationInvites(ctx context.Context, req *adminv1.ListOrganizationInvitesRequest) (*adminv1.ListOrganizationInvitesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	// get pending user invites for this org
	userInvites, err := s.admin.DB.FindOrganizationInvites(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nextToken := ""
	if len(userInvites) >= pageSize {
		nextToken = marshalPageToken(userInvites[len(userInvites)-1].Email)
	}

	invitesDtos := make([]*adminv1.UserInvite, len(userInvites))
	for i, invite := range userInvites {
		invitesDtos[i] = inviteToPB(invite)
	}

	return &adminv1.ListOrganizationInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) AddOrganizationMemberUser(ctx context.Context, req *adminv1.AddOrganizationMemberUserRequest) (*adminv1.AddOrganizationMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org members")
	}

	count, err := s.admin.DB.CountInvitesForOrganization(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if org.QuotaOutstandingInvites >= 0 && count >= org.QuotaOutstandingInvites {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can at most have %d outstanding invitations", org.QuotaOutstandingInvites)
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var invitedByUserID, invitedByName string
	if claims.OwnerType() == auth.OwnerTypeUser {
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		invitedByUserID = user.ID
		invitedByName = user.DisplayName
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Invite user to join org
		err := s.admin.DB.InsertOrganizationInvite(ctx, &database.InsertOrganizationInviteOptions{
			Email:     req.Email,
			InviterID: invitedByUserID,
			OrgID:     org.ID,
			RoleID:    role.ID,
		})
		// continue sending an email if the user already exists
		if err != nil && !errors.Is(err, database.ErrNotUnique) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Send invitation email
		err = s.admin.Email.SendOrganizationInvite(&email.OrganizationInvite{
			ToEmail:       req.Email,
			ToName:        "",
			AcceptURL:     s.admin.URLs.WithCustomDomain(org.CustomDomain).OrganizationInviteAccept(org.Name),
			OrgName:       org.Name,
			RoleName:      role.Name,
			InvitedByName: invitedByName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &adminv1.AddOrganizationMemberUserResponse{
			PendingSignup: true,
		}, nil
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	errored := false
	err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
	// continue sending an email if the user already exists
	if err != nil {
		errored = true
		if !errors.Is(err, database.ErrNotUnique) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if !errored {
		// if previous statement errored we cannot continue with this since transaction would be invalid
		err = s.admin.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			if !errors.Is(err, database.ErrNotUnique) {
				return nil, status.Error(codes.Internal, err.Error())
			}
			// If the user is already in the all user group, we can ignore the error
		}

		err = tx.Commit()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = s.admin.Email.SendOrganizationAddition(&email.OrganizationAddition{
		ToEmail:       req.Email,
		ToName:        "",
		OpenURL:       s.admin.URLs.WithCustomDomain(org.CustomDomain).Organization(org.Name),
		OrgName:       org.Name,
		RoleName:      role.Name,
		InvitedByName: invitedByName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddOrganizationMemberUserResponse{
		PendingSignup: false,
	}, nil
}

func (s *Server) RemoveOrganizationMemberUser(ctx context.Context, req *adminv1.RemoveOrganizationMemberUserRequest) (*adminv1.RemoveOrganizationMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.Bool("args.keep_project_roles", req.KeepProjectRoles),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Only admins can remove pending invites.
		// NOTE: If we change invites to accept/decline (instead of auto-accept on signup), we need to revisit this.
		claims := auth.GetClaims(ctx)
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
		}

		// Check if there is a pending invite
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		err = s.admin.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &adminv1.RemoveOrganizationMemberUserResponse{}, nil
	}

	// The caller must either have ManageOrgMembers permission or be the user being removed.
	claims := auth.GetClaims(ctx)
	isManager := claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers
	isSelf := claims.OwnerType() == auth.OwnerTypeUser && claims.OwnerID() == user.ID
	if !isManager && !isSelf {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	if org.BillingEmail == user.Email {
		return nil, status.Error(codes.InvalidArgument, "this user is the billing email for the organization, please update the billing email before removing")
	}

	// Check that the user is not the last admin
	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(users) == 1 && users[0].ID == user.ID {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last admin member")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user groups of the org
	err = s.admin.DB.DeleteUsergroupsMemberUser(ctx, org.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// delete from projects if KeepProjectRoles flag is set
	if !req.KeepProjectRoles {
		err = s.admin.DB.DeleteAllProjectMemberUserForOrganization(ctx, org.ID, user.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveOrganizationMemberUserResponse{}, nil
}

func (s *Server) SetOrganizationMemberUserRole(ctx context.Context, req *adminv1.SetOrganizationMemberUserRoleRequest) (*adminv1.SetOrganizationMemberUserRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.DB.UpdateOrganizationInviteRole(ctx, invite.ID, role.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &adminv1.SetOrganizationMemberUserRoleResponse{}, nil
	}

	// Check if the user is the last owner
	if role.Name != database.OrganizationRoleNameAdmin {
		adminRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
		if err != nil {
			panic(err)
		}
		// TODO optimize this, may be extract roles during auth token validation
		//  and store as part of the claims and fetch admins only if the user is an admin
		users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, adminRole.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if len(users) == 1 && users[0].ID == user.ID {
			return nil, status.Error(codes.InvalidArgument, "cannot change role of the last owner")
		}
	}

	err = s.admin.DB.UpdateOrganizationMemberUserRole(ctx, org.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.SetOrganizationMemberUserRoleResponse{}, nil
}

func (s *Server) LeaveOrganization(ctx context.Context, req *adminv1.LeaveOrganizationRequest) (*adminv1.LeaveOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(err)
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if org.BillingEmail == user.Email {
		return nil, status.Error(codes.InvalidArgument, "this user is the billing email for the organization, please update the billing email before leaving")
	}

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(users) == 1 && users[0].ID == claims.OwnerID() {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user groups of the org
	err = s.admin.DB.DeleteUsergroupsMemberUser(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.LeaveOrganizationResponse{}, nil
}

func (s *Server) CreateWhitelistedDomain(ctx context.Context, req *adminv1.CreateWhitelistedDomainRequest) (*adminv1.CreateWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.domain", req.Domain),
		attribute.String("args.role", req.Role),
	)

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Superuser(ctx) {
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
			return nil, status.Error(codes.PermissionDenied, "only org admins can add whitelisted domain")
		}
		// check if the user's domain matches the whitelist domain
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !strings.HasSuffix(user.Email, "@"+req.Domain) {
			return nil, status.Error(codes.PermissionDenied, "Domain name doesn’t match verified email domain. Please contact Rill support.")
		}

		if publicemail.IsPublic(req.Domain) {
			return nil, status.Errorf(codes.InvalidArgument, "Public Domain %s cannot be whitelisted", req.Domain)
		}
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// find existing users belonging to the whitelisted domain to the org
	users, err := s.admin.DB.FindUsersByEmailPattern(ctx, "%@"+req.Domain, "", math.MaxInt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// filter out users who are already members of the org
	newUsers := make([]*database.User, 0)
	for _, user := range users {
		// check if user is already a member of the org
		exists, err := s.admin.DB.CheckUserIsAnOrganizationMember(ctx, user.ID, org.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !exists {
			newUsers = append(newUsers, user)
		}
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = s.admin.DB.InsertOrganizationWhitelistedDomain(ctx, &database.InsertOrganizationWhitelistedDomainOptions{
		OrgID:     org.ID,
		OrgRoleID: role.ID,
		Domain:    req.Domain,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotUnique) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, user := range newUsers {
		err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// add to all user group
		err = s.admin.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateWhitelistedDomainResponse{}, nil
}

func (s *Server) RemoveWhitelistedDomain(ctx context.Context, req *adminv1.RemoveWhitelistedDomainRequest) (*adminv1.RemoveWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.domain", req.Domain),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can remove whitelisted domain")
	}

	invite, err := s.admin.DB.FindOrganizationWhitelistedDomain(ctx, org.ID, req.Domain)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "whitelist not found for org %q and domain %q", org.Name, req.Domain)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.DeleteOrganizationWhitelistedDomain(ctx, invite.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveWhitelistedDomainResponse{}, nil
}

func (s *Server) ListWhitelistedDomains(ctx context.Context, req *adminv1.ListWhitelistedDomainsRequest) (*adminv1.ListWhitelistedDomainsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can list whitelisted domain")
	}

	whitelistedDomains, err := s.admin.DB.FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	whitelistedDomainDtos := make([]*adminv1.WhitelistedDomain, len(whitelistedDomains))
	for i, whitelistedDomain := range whitelistedDomains {
		whitelistedDomainDtos[i] = whitelistedDomainToPB(whitelistedDomain)
	}

	return &adminv1.ListWhitelistedDomainsResponse{
		Domains: whitelistedDomainDtos,
	}, nil
}

func (s *Server) SudoUpdateOrganizationQuotas(ctx context.Context, req *adminv1.SudoUpdateOrganizationQuotasRequest) (*adminv1.SudoUpdateOrganizationQuotasResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.OrgName))
	if req.Projects != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.projects", int(*req.Projects)))
	}
	if req.Deployments != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.deployments", int(*req.Deployments)))
	}
	if req.SlotsTotal != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_total", int(*req.SlotsTotal)))
	}
	if req.SlotsPerDeployment != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_per_deployment", int(*req.SlotsPerDeployment)))
	}
	if req.OutstandingInvites != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.outstanding_invites", int(*req.OutstandingInvites)))
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage quotas")
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
		QuotaProjects:                       int(valOrDefault(req.Projects, int32(org.QuotaProjects))),
		QuotaDeployments:                    int(valOrDefault(req.Deployments, int32(org.QuotaDeployments))),
		QuotaSlotsTotal:                     int(valOrDefault(req.SlotsTotal, int32(org.QuotaSlotsTotal))),
		QuotaSlotsPerDeployment:             int(valOrDefault(req.SlotsPerDeployment, int32(org.QuotaSlotsPerDeployment))),
		QuotaOutstandingInvites:             int(valOrDefault(req.OutstandingInvites, int32(org.QuotaOutstandingInvites))),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(req.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	}

	updatedOrg, err := s.admin.DB.UpdateOrganization(ctx, org.ID, opts)
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateOrganizationQuotasResponse{
		Organization: organizationToDTO(updatedOrg),
	}, nil
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

func (s *Server) SudoUpdateOrganizationCustomDomain(ctx context.Context, req *adminv1.SudoUpdateOrganizationCustomDomainRequest) (*adminv1.SudoUpdateOrganizationCustomDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Name),
		attribute.String("args.custom_domain", req.CustomDomain),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage custom domains")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        req.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateOrganizationCustomDomainResponse{
		Organization: organizationToDTO(org),
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
	case database.BillingWarningTypeTrialEnding:
		return adminv1.BillingWarningType_BILLING_WARNING_TYPE_TRIAL_ENDING
	default:
		return adminv1.BillingWarningType_BILLING_WARNING_TYPE_UNSPECIFIED
	}
}

func dtoBillingWarningTypeToDB(t adminv1.BillingWarningType) (database.BillingWarningType, error) {
	switch t {
	case adminv1.BillingWarningType_BILLING_WARNING_TYPE_UNSPECIFIED:
		return database.BillingWarningTypeUnspecified, nil
	case adminv1.BillingWarningType_BILLING_WARNING_TYPE_TRIAL_ENDING:
		return database.BillingWarningTypeTrialEnding, nil
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

func organizationToDTO(o *database.Organization) *adminv1.Organization {
	return &adminv1.Organization{
		Id:           o.ID,
		Name:         o.Name,
		DisplayName:  o.DisplayName,
		Description:  o.Description,
		CustomDomain: o.CustomDomain,
		Quotas: &adminv1.OrganizationQuotas{
			Projects:                       int32(o.QuotaProjects),
			Deployments:                    int32(o.QuotaDeployments),
			SlotsTotal:                     int32(o.QuotaSlotsTotal),
			SlotsPerDeployment:             int32(o.QuotaSlotsPerDeployment),
			OutstandingInvites:             int32(o.QuotaOutstandingInvites),
			StorageLimitBytesPerDeployment: o.QuotaStorageLimitBytesPerDeployment,
		},
		BillingCustomerId: o.BillingCustomerID,
		PaymentCustomerId: o.PaymentCustomerID,
		BillingEmail:      o.BillingEmail,
		CreatedOn:         timestamppb.New(o.CreatedOn),
		UpdatedOn:         timestamppb.New(o.UpdatedOn),
	}
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

func valOrEmptyString(v *int) string {
	if v != nil {
		return strconv.Itoa(*v)
	}
	return ""
}

func val64OrEmptyString(v *int64) string {
	if v != nil {
		return strconv.FormatInt(*v, 10)
	}
	return ""
}

func whitelistedDomainToPB(a *database.OrganizationWhitelistedDomainWithJoinedRoleNames) *adminv1.WhitelistedDomain {
	return &adminv1.WhitelistedDomain{
		Domain: a.Domain,
		Role:   a.RoleName,
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
