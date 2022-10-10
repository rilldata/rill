package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/organization"

	_ "github.com/mattn/go-sqlite3"
)

// (GET /v1/organizations)
func (s *Server) FindOrganizations(ctx echo.Context) error {
	orgs, err := QueryOrganizations(ctx.Request().Context(), s.client)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Organization, len(orgs))
	for i, org := range orgs {
		dtos[i] = orgToDTO(org)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/organizations)
func (s *Server) CreateOrganization(ctx echo.Context) error {
	var dto api.CreateOrganizationJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	org, err := s.client.Organization.
		Create().
		SetName(dto.Name).
		SetDescription(*dto.Description).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed creating Organization: %w", err)
	}

	return ctx.JSON(http.StatusCreated, orgToDTO(org))
}

// (DELETE /v1/organizations/{name})
func (s *Server) DeleteOrganization(ctx echo.Context, name string) error {
	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.client.Organization.DeleteOne(org).Exec(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, orgToDTO(org))
}

// (GET /v1/organizations/{name})
func (s *Server) FindOrganization(ctx echo.Context, name string) error {
	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, orgToDTO(org))
}

// (PUT /organizations/{name})
func (s *Server) UpdateOrganization(ctx echo.Context, name string) error {
	var dto api.UpdateOrganizationJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	orgNew, err := s.client.Organization.
		UpdateOne(org).
		// RemoveUsers(). //Do we need to removed edges in case of update?
		SetDescription(*dto.Description).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed updating Organization: %w", err)
	}

	return ctx.JSON(http.StatusOK, orgToDTO(orgNew))
}

// (GET /v1/organizations/{organization}/users)
func (s *Server) ListUsers(ctx echo.Context, orgName string) error {
	users, err := QueryUsersByOrgName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.User, len(users))
	for i, user := range users {
		dtos[i] = userToDTO(user)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/organizations/{organization}/users)
func (s *Server) AddUser(ctx echo.Context, orgName string) error {
	var dto api.AddUserJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := QueryUserByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newOrg, err := org.Update().AddUsers(user).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	users, err := newOrg.QueryUsers().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.User, len(users))
	for i, user := range users {
		dtos[i] = userToDTO(user)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (PUT /v1/organizations/{organization}/users)
func (s *Server) RemoveUser(ctx echo.Context, orgName string) error {
	var dto api.RemoveUserJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := QueryUserByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newOrg, err := org.Update().RemoveUsers(user).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	users, err := newOrg.QueryUsers().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.User, len(users))
	for i, user := range users {
		dtos[i] = userToDTO(user)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

func orgToDTO(o *ent.Organization) *api.Organization {
	return &api.Organization{
		Id:          fmt.Sprint(o.ID),
		Name:        o.Name,
		Description: stringToPtr(o.Description),
		CreatedOn:   types.Date{Time: o.CreatedOn},
		UpdatedOn:   types.Date{Time: o.CreatedOn},
	}
}

func QueryOrganizations(ctx context.Context, client *ent.Client) ([]*ent.Organization, error) {
	u, err := client.Organization.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Organization: %w", err)
	}

	return u, nil
}

func QueryOrganizationByName(ctx context.Context, client *ent.Client, name string) (*ent.Organization, error) {
	org, err := client.Organization.
		Query().
		Where(organization.NameEQ(name)).
		// `Only` fails if no org found,
		// or more than 1 org returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Organization: %w", err)
	}

	return org, nil
}

func QueryUsersByOrgName(ctx context.Context, client *ent.Client, orgName string) ([]*ent.User, error) {
	org, err := QueryOrganizationByName(ctx, client, orgName)
	if err != nil {
		return nil, err
	}

	u, err := org.QueryUsers().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying User: %w", err)
	}
	return u, nil
}
