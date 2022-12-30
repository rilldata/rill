package server

import (
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/admin/api"
)

// Sends an error as JSON
// TODO: Replace with a custom error handler on echo.
func sendError(ctx echo.Context, code int, message string) error {
	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, apiErr)
	return err
}

// stringFromPtr dereferences ptr and returns it, defaulting to an empty string if ptr is nil.
func stringFromPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// stringToPtr returns nil if s is empty, otherwise returns &s.
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
