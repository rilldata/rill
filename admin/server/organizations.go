package server

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/admin/api"
	"github.com/rilldata/rill/admin/database"
)

// (GET /v1/organizations)
func (s *Server) FindOrganizations(ctx echo.Context) error {
	orgs, err := s.db.FindOrganizations(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Organization, len(orgs))
	for i, org := range orgs {
		dtos[i] = orgToDTO(org)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /organizations)
func (s *Server) CreateOrganization(ctx echo.Context) error {
	var dto api.CreateOrganizationJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	org, err := s.db.CreateOrganization(ctx.Request().Context(), dto.Name, stringFromPtr(dto.Description))
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusCreated, orgToDTO(org))
}

// (DELETE /organizations/{name})
func (s *Server) DeleteOrganization(ctx echo.Context, name string) error {
	err := s.db.DeleteOrganization(ctx.Request().Context(), name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

// (GET /organizations/{name})
func (s *Server) FindOrganization(ctx echo.Context, name string) error {
	org, err := s.db.FindOrganizationByName(ctx.Request().Context(), name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, "not found")
		}
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

	org, err := s.db.UpdateOrganization(ctx.Request().Context(), name, stringFromPtr(dto.Description))
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, orgToDTO(org))
}

func orgToDTO(o *database.Organization) *api.Organization {
	return &api.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Description: stringToPtr(o.Description),
		CreatedOn:   types.Date{Time: o.CreatedOn},
		UpdatedOn:   types.Date{Time: o.CreatedOn},
	}
}
