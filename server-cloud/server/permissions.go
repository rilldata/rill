package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/permission"
)

// (GET /v1/permissions)
func (s *Server) FindPermissions(ctx echo.Context) error {
	permissions, err := QueryPermissions(ctx.Request().Context(), s.client)
	if err != nil {
		return fmt.Errorf("failed querying users: %w", err)
	}

	dtos := make([]*api.Permission, len(permissions))
	for i, permission := range permissions {
		dtos[i] = permissionToDTO(permission)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/permissions)
func (s *Server) CreatePermission(ctx echo.Context) error {
	var dto api.CreatePermissionJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	permission, err := s.client.Permission.
		Create().
		SetName(dto.Name).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed creating Permission: %w", err)
	}

	return ctx.JSON(http.StatusCreated, permissionToDTO(permission))
}

// (GET /v1/permission/{name})
func (s *Server) FindPermission(ctx echo.Context, name string) error {
	permission, err := QueryPermissionByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, permissionToDTO(permission))
}

// (DELETE /v1/permission/{name})
func (s *Server) DeletePermission(ctx echo.Context, name string) error {
	permission, err := QueryPermissionByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.client.Permission.DeleteOne(permission).Exec(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, permissionToDTO(permission))
}

func permissionToDTO(o *ent.Permission) *api.Permission {
	return &api.Permission{
		Id:        fmt.Sprint(o.ID),
		Name:      o.Name,
		CreatedOn: types.Date{Time: o.CreatedOn},
		UpdatedOn: types.Date{Time: o.CreatedOn},
	}
}

func QueryPermissions(ctx context.Context, client *ent.Client) ([]*ent.Permission, error) {
	permission, err := client.Permission.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Permissions: %w", err)
	}

	return permission, nil
}

func QueryPermissionByName(ctx context.Context, client *ent.Client, name string) (*ent.Permission, error) {
	permission, err := client.Permission.
		Query().
		Where(permission.NameEQ(name)).
		// `Only` fails if no permission found,
		// or more than 1 permission returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Permission: %w", err)
	}

	return permission, nil
}
