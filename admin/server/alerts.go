package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

func (s *Server) GetAlertMeta(ctx context.Context, req *adminv1.GetAlertMetaRequest) (*adminv1.GetAlertMetaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.alert", req.Alert),
		attribute.Bool("args.query_for", req.GetQueryFor() != nil),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read alert meta")
	}

	if proj.ProdBranch != req.Branch {
		return nil, status.Error(codes.InvalidArgument, "branch not found")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var attr map[string]any
	if req.QueryFor != nil {
		// TODO: This logic is duplicated in GetDeploymentCredentials. We should extract to a single implementation.
		switch forVal := req.QueryFor.(type) {
		case *adminv1.GetAlertMetaRequest_QueryForUserId:
			attr, err = s.getAttributesFor(ctx, forVal.QueryForUserId, proj.OrganizationID, proj.ID)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		case *adminv1.GetAlertMetaRequest_QueryForUserEmail:
			user, err := s.admin.DB.FindUserByEmail(ctx, forVal.QueryForUserEmail)
			// if email is not found in the database, we assume it is a non-admin user
			if errors.Is(err, database.ErrNotFound) {
				attr = map[string]any{
					"email":  forVal.QueryForUserEmail,
					"domain": forVal.QueryForUserEmail[strings.LastIndex(forVal.QueryForUserEmail, "@")+1:],
					"admin":  false,
				}
				break
			}
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			attr, err = s.getAttributesFor(ctx, user.ID, proj.OrganizationID, proj.ID)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
		}
	}

	var attrPB *structpb.Struct
	if attr != nil {
		attrPB, err = structpb.NewStruct(attr)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &adminv1.GetAlertMetaResponse{
		OpenUrl:            s.urls.alertOpen(org.Name, proj.Name, req.Alert),
		EditUrl:            s.urls.alertEdit(org.Name, proj.Name, req.Alert),
		QueryForAttributes: attrPB,
	}, nil
}

func (s *Server) CreateAlert(ctx context.Context, req *adminv1.CreateAlertRequest) (*adminv1.CreateAlertResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.CreateAlerts {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can create alerts")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	name, err := s.generateAlertName(ctx, depl, req.Options.Title)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	data, err := s.yamlForManagedAlert(req.Options, name, claims.OwnerID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate alert YAML: %s", err.Error())
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID: proj.ID,
		Branch:    proj.ProdBranch,
		Path:      virtualFilePathForManagedAlert(name),
		Data:      data,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitResource(ctx, depl, name, runtime.ResourceKindAlert)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for alert to be created")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile alert: %s", err.Error())
	}

	return &adminv1.CreateAlertResponse{
		Name: name,
	}, nil
}

func (s *Server) EditAlert(ctx context.Context, req *adminv1.EditAlertRequest) (*adminv1.EditAlertResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupAlert(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get alert: %s", err.Error())
	}
	annotations := parseAlertAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit alert because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && annotations.AdminOwnerUserID == claims.OwnerID()
	if !permissions.ManageAlerts && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit alert")
	}

	data, err := s.yamlForManagedAlert(req.Options, req.Name, annotations.AdminOwnerUserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate alert YAML: %s", err.Error())
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID: proj.ID,
		Branch:    proj.ProdBranch,
		Path:      virtualFilePathForManagedAlert(req.Name),
		Data:      data,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindAlert)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for alert to be updated")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile alert: %s", err.Error())
	}

	return &adminv1.EditAlertResponse{}, nil
}

func (s *Server) UnsubscribeAlert(ctx context.Context, req *adminv1.UnsubscribeAlertRequest) (*adminv1.UnsubscribeAlertResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupAlert(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get alert: %s", err.Error())
	}
	annotations := parseAlertAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit alert because it was not created from the UI")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can unsubscribe from alerts")
	}
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	opts, err := recreateAlertOptionsFromSpec(spec)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to recreate alert options: %s", err.Error())
	}

	found := false
	for idx, email := range opts.Recipients {
		if strings.EqualFold(user.Email, email) {
			opts.Recipients = slices.Delete(opts.Recipients, idx, idx+1)
			found = true
			break
		}
	}

	if !found {
		return nil, status.Error(codes.InvalidArgument, "user is not subscribed to alert")
	}

	if len(opts.Recipients) == 0 {
		err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, proj.ProdBranch, virtualFilePathForManagedAlert(req.Name))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
		}
	} else {
		data, err := s.yamlForManagedAlert(opts, req.Name, annotations.AdminOwnerUserID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to generate alert YAML: %s", err.Error())
		}

		err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID: proj.ID,
			Branch:    proj.ProdBranch,
			Path:      virtualFilePathForManagedAlert(req.Name),
			Data:      data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
		}
	}

	err = s.admin.TriggerReconcileAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindAlert)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for alert to be updated")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile alert: %s", err.Error())
	}

	return &adminv1.UnsubscribeAlertResponse{}, nil
}

func (s *Server) DeleteAlert(ctx context.Context, req *adminv1.DeleteAlertRequest) (*adminv1.DeleteAlertResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupAlert(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get alert: %s", err.Error())
	}
	annotations := parseAlertAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit alert because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && annotations.AdminOwnerUserID == claims.OwnerID()
	if !permissions.ManageAlerts && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit alert")
	}

	err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, proj.ProdBranch, virtualFilePathForManagedAlert(req.Name))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindAlert)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for alert to be deleted")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile alert: %s", err.Error())
	}

	return &adminv1.DeleteAlertResponse{}, nil
}

func (s *Server) GenerateAlertYAML(ctx context.Context, req *adminv1.GenerateAlertYAMLRequest) (*adminv1.GenerateAlertYAMLResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	data, err := s.yamlForCommittedAlert(req.Options)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate alert YAML: %s", err.Error())
	}

	return &adminv1.GenerateAlertYAMLResponse{
		Yaml: string(data),
	}, nil
}

func (s *Server) GetAlertYAML(ctx context.Context, req *adminv1.GetAlertYAMLRequest) (*adminv1.GetAlertYAMLResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupAlert(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get alert: %s", err.Error())
	}
	annotations := parseAlertAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit alert because it was not created from the UI")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can view the alert yaml")
	}

	opts, err := recreateAlertOptionsFromSpec(spec)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to recreate alert options: %s", err.Error())
	}

	data, err := s.yamlForCommittedAlert(opts)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate alert YAML: %s", err.Error())
	}

	return &adminv1.GetAlertYAMLResponse{
		Yaml: string(data),
	}, nil
}

func (s *Server) yamlForManagedAlert(opts *adminv1.AlertOptions, alertName, ownerUserID string) ([]byte, error) {
	res := alertYAML{}
	res.Kind = "alert"
	res.Title = opts.Title
	res.Intervals.Duration = opts.IntervalDuration
	res.Query.Name = opts.QueryName
	res.Query.ArgsJSON = opts.QueryArgsJson
	// hard code the user id to run for. this avoids exposing data through alert creation
	res.Query.For.UserID = ownerUserID
	res.Email.Recipients = opts.Recipients
	res.Email.Renotify = opts.EmailRenotify
	res.Email.RenotifyAfter = opts.EmailRenotifyAfterSeconds
	res.Annotations.AdminOwnerUserID = ownerUserID
	res.Annotations.AdminManaged = true
	res.Annotations.AdminNonce = time.Now().Format(time.RFC3339Nano)
	return yaml.Marshal(res)
}

func (s *Server) yamlForCommittedAlert(opts *adminv1.AlertOptions) ([]byte, error) {
	// Format args as pretty YAML
	var args map[string]interface{}
	if opts.QueryArgsJson != "" {
		err := json.Unmarshal([]byte(opts.QueryArgsJson), &args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse queryArgsJSON: %w", err)
		}
	}

	res := alertYAML{}
	res.Kind = "alert"
	res.Title = opts.Title
	res.Intervals.Duration = opts.IntervalDuration
	res.Query.Name = opts.QueryName
	res.Query.Args = args
	res.Email.Recipients = opts.Recipients
	res.Email.Renotify = opts.EmailRenotify
	res.Email.RenotifyAfter = opts.EmailRenotifyAfterSeconds
	return yaml.Marshal(res)
}

// generateAlertName generates a random alert name with the title as a seed.
// Example: "My alert!" -> "my-alert-5b3f7e1a".
// It verifies that the name is not taken (the random component makes any collision unlikely, but we check to be sure).
func (s *Server) generateAlertName(ctx context.Context, depl *database.Deployment, title string) (string, error) {
	for i := 0; i < 5; i++ {
		name := randomAlertName(title)

		_, err := s.admin.LookupAlert(ctx, depl, name)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				// Success! Name isn't taken
				return name, nil
			}
			return "", fmt.Errorf("failed to check alert name: %w", err)
		}
	}

	// Fail-safe in case all names we tried were taken
	return uuid.New().String(), nil
}

var alertNameToDashCharsRegexp = regexp.MustCompile(`[ _]+`)

var alertNameExcludeCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func randomAlertName(title string) string {
	name := alertNameToDashCharsRegexp.ReplaceAllString(title, "-")
	name = alertNameExcludeCharsRegexp.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	name = strings.Trim(name, "-")
	if name == "" {
		name = uuid.New().String()
	} else {
		name = name + "-" + uuid.New().String()[0:8]
	}
	return name
}

func recreateAlertOptionsFromSpec(spec *runtimev1.AlertSpec) (*adminv1.AlertOptions, error) {
	opts := &adminv1.AlertOptions{}
	opts.Title = spec.Title
	opts.IntervalDuration = spec.IntervalsIsoDuration
	opts.QueryName = spec.QueryName
	opts.QueryArgsJson = spec.QueryArgsJson
	opts.Recipients = spec.EmailRecipients
	opts.EmailRenotify = spec.EmailRenotify
	opts.EmailRenotifyAfterSeconds = spec.EmailRenotifyAfterSeconds
	return opts, nil
}

// alertYAML is derived from rillv1.AlertYAML, but adapted for generating (as opposed to parsing) the alert YAML.
type alertYAML struct {
	Kind      string `yaml:"kind"`
	Title     string `yaml:"title"`
	Intervals struct {
		Duration string `yaml:"duration"`
	} `yaml:"intervals"`
	Timeout string `yaml:"timeout"`
	Query   struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args,omitempty"`
		ArgsJSON string         `yaml:"args_json,omitempty"`
		For      struct {
			UserID string `yaml:"user_id"`
		} `yaml:"for"`
	} `yaml:"query"`
	Email struct {
		Recipients    []string `yaml:"recipients"`
		Renotify      bool     `yaml:"renotify"`
		RenotifyAfter uint32   `yaml:"renotify_after"`
	} `yaml:"email"`
	Annotations alertAnnotations `yaml:"annotations,omitempty"`
}

type alertAnnotations struct {
	AdminOwnerUserID string `yaml:"admin_owner_user_id"`
	AdminManaged     bool   `yaml:"admin_managed"`
	AdminNonce       string `yaml:"admin_nonce"` // To ensure spec version gets updated on writes, to enable polling in TriggerReconcileAndAwaitAlert
}

func parseAlertAnnotations(annotations map[string]string) alertAnnotations {
	if annotations == nil {
		return alertAnnotations{}
	}

	res := alertAnnotations{}
	res.AdminOwnerUserID = annotations["admin_owner_user_id"]
	res.AdminManaged, _ = strconv.ParseBool(annotations["admin_managed"])
	res.AdminNonce = annotations["admin_nonce"]

	return res
}

func virtualFilePathForManagedAlert(name string) string {
	return path.Join("alerts", name+".yaml")
}
