package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/role"
)

// (GET /v1/roles)
func (s *Server) FindRoles(ctx echo.Context) error {
	roles, err := QueryRoles(ctx.Request().Context(), s.client)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Role, len(roles))
	for i, role := range roles {
		dtos[i] = roleToDTO(role)
	}

	LoadRoles(ctx.Request().Context(), s.client)

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/roles)
func (s *Server) CreateRole(ctx echo.Context) error {
	var dto api.CreateRoleJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	role, err := s.client.Role.
		Create().
		SetName(dto.Name).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed creating Role: %w", err)
	}

	return ctx.JSON(http.StatusCreated, roleToDTO(role))
}

// (GET /v1/roles/{name})
func (s *Server) FindRole(ctx echo.Context, name string) error {
	role, err := QueryRoleByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, roleToDTO(role))
}

// (DELETE /v1/roles/{name})
func (s *Server) DeleteRole(ctx echo.Context, name string) error {
	role, err := QueryRoleByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.client.Role.DeleteOne(role).Exec(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, roleToDTO(role))
}

// (GET /v1/roles/{name}/permissions)
func (s *Server) ListPermissions(ctx echo.Context, roleName string) error {
	role, err := QueryRoleByName(ctx.Request().Context(), s.client, roleName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	permissions, err := role.QueryPermission().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Errorf("failed querying Permissions: %w", err).Error())
	}

	dtos := make([]*api.Permission, len(permissions))
	for i, permission := range permissions {
		dtos[i] = permissionToDTO(permission)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/roles/{name}/permissions)
func (s *Server) AddPermission(ctx echo.Context, roleName string) error {
	var dto api.AddPermissionJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	role, err := QueryRoleByName(ctx.Request().Context(), s.client, roleName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	permission, err := QueryPermissionByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newRole, err := s.client.Role.UpdateOne(role).AddPermission(permission).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Errorf("failed adding permission to Role: %w", err).Error())
	}

	permissions, err := newRole.QueryPermission().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Permission, len(permissions))
	for i, permission := range permissions {
		dtos[i] = permissionToDTO(permission)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (PUT /v1/roles/{name}/permissions)
func (s *Server) RemovePermission(ctx echo.Context, roleName string) error {
	var dto api.RemovePermissionJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	role, err := QueryRoleByName(ctx.Request().Context(), s.client, roleName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	permission, err := QueryPermissionByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newRole, err := role.Update().RemovePermission(permission).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Errorf("failed adding permission to Role: %w", err).Error())
	}

	permissions, err := newRole.QueryPermission().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Permission, len(permissions))
	for i, permission := range permissions {
		dtos[i] = permissionToDTO(permission)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

func roleToDTO(o *ent.Role) *api.Role {
	return &api.Role{
		Id:        fmt.Sprint(o.ID),
		Name:      o.Name,
		CreatedOn: types.Date{Time: o.CreatedOn},
		UpdatedOn: types.Date{Time: o.CreatedOn},
	}
}

func QueryRoles(ctx context.Context, client *ent.Client) ([]*ent.Role, error) {
	roles, err := client.Role.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Role: %w", err)
	}

	return roles, nil
}

func QueryRoleByName(ctx context.Context, client *ent.Client, name string) (*ent.Role, error) {
	role, err := client.Role.
		Query().
		Where(role.NameEQ(name)).
		// `Only` fails if no role found,
		// or more than 1 role returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Role: %w", err)
	}

	return role, nil
}

func LoadRoles(ctx context.Context, client *ent.Client) (interface{}, error) {
	rolesMap := make(map[string][]string)

	entRoles, err := QueryRoles(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, role := range entRoles {
		permissions := role.QueryPermission().AllX(ctx)
		var permissionList []string
		for _, permission := range permissions {
			permissionList = append(permissionList, permission.Name)

		}
		rolesMap[role.Name] = permissionList
	}

	return rolesMap, nil
}
