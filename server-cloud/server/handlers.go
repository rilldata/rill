package server

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/database"
)

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
		if err == database.ErrNotFound {
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

// Sends an error as JSON
// TODO: Replace with a custom error handler on echo
func sendError(ctx echo.Context, code int, message string) error {
	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, apiErr)
	return err
}

// stringFromPtr dereferences ptr and returns it, defaulting to an empty string if ptr is nil
func stringFromPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// stringToPtr returns nil if s is empty, otherwise returns &s
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
