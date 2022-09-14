package server

import (
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
)

// (POST /organizations)
func (s *Server) CreateOrganization(ctx echo.Context, params api.CreateOrganizationParams) error {
	return sendError(ctx, 400, "not implemented")
	// return ctx.JSON(http.StatusOK, result)
}

// (DELETE /organizations/{name})
func (s *Server) DeleteOrganization(ctx echo.Context, name string) error {
	return sendError(ctx, 400, "not implemented")
}

// (GET /organizations/{name})
func (s *Server) FindOrganization(ctx echo.Context, name string) error {
	return sendError(ctx, 400, "not implemented")
}

// (PUT /organizations/{name})
func (s *Server) UpdateOrganization(ctx echo.Context, name string, params api.UpdateOrganizationParams) error {
	return sendError(ctx, 400, "not implemented")
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendError(ctx echo.Context, code int, message string) error {
	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, apiErr)
	return err
}
