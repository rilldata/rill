package server

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userForPrimaryDeployment checks that the caller is the project's primary deployment, and returns the user it acts for with their permissions on the project.
// The deployment is trusted to pass the user of its session, as for GetAlertMeta, GetReportMeta and GetConnectorOAuthToken.
// Other deployments are rejected, because alerts and reports created from the UI live in the prod environment, which only the primary deployment reads.
func (s *Server) userForPrimaryDeployment(ctx context.Context, proj *database.Project, userID string) (*database.User, *adminv1.ProjectPermissions, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeDeployment || proj.PrimaryDeploymentID == nil || *proj.PrimaryDeploymentID != claims.OwnerID() {
		return nil, nil, status.Error(codes.PermissionDenied, "only the project's primary deployment can act for a user")
	}

	// Only users, not services, magic tokens or embed users, own the resources this creates.
	if _, err := uuid.Parse(userID); err != nil {
		return nil, nil, status.Error(codes.PermissionDenied, "can only act for a user")
	}
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil, status.Error(codes.PermissionDenied, "can only act for a user")
		}
		return nil, nil, err
	}

	orgPerms, err := s.admin.OrganizationPermissionsForUser(ctx, proj.OrganizationID, user.ID)
	if err != nil {
		return nil, nil, err
	}
	projPerms, err := s.admin.ProjectPermissionsForUser(ctx, proj.ID, user.ID, orgPerms)
	if err != nil {
		return nil, nil, err
	}
	return user, projPerms, nil
}

// recipientsForUser checks the recipients of an alert or report that the AI creates for a user, and returns its email recipients.
// The AI can be steered by the data and tools it reads, so what it creates can only reach members of the project:
// the email recipients default to the user, other email recipients and Slack users must be members of the project,
// and Slack channels and webhooks, whose readers can't be checked, are rejected.
func (s *Server) recipientsForUser(ctx context.Context, proj *database.Project, user *database.User, emails, slackUsers, slackChannels, slackWebhooks []string) ([]string, error) {
	if len(slackChannels) > 0 || len(slackWebhooks) > 0 {
		return nil, status.Error(codes.InvalidArgument, "the AI can't send to Slack channels or webhooks, since it can't check who reads them")
	}
	if len(emails) == 0 && len(slackUsers) == 0 {
		return []string{user.Email}, nil
	}

	for _, recipient := range append(append([]string{}, emails...), slackUsers...) {
		if strings.EqualFold(recipient, user.Email) {
			continue
		}
		_, memberID, err := s.getAttributesForProjectMember(ctx, recipient, proj.OrganizationID, proj.ID)
		if err != nil {
			return nil, err
		}
		if memberID == "" {
			return nil, status.Errorf(codes.InvalidArgument, "%s is not a member of the project, and the AI can only send to project members", recipient)
		}
	}
	return emails, nil
}
