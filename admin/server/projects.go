package server

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/admin/api"
	"github.com/rilldata/rill/admin/database"
)

// (GET /v1/organizations/{organization}/projects)
func (s *Server) FindProjects(ctx echo.Context, organization string) error {
	projs, err := s.db.FindProjects(ctx.Request().Context(), organization)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Project, len(projs))
	for i, proj := range projs {
		dtos[i] = projToDTO(proj)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (GET /v1/organizations/{organization}/project/{name})
func (s *Server) FindProject(ctx echo.Context, organization, name string) error {
	proj, err := s.db.FindProjectByName(ctx.Request().Context(), organization, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, projToDTO(proj))
}

// (POST /v1/organizations/{organization}/projects)
func (s *Server) CreateProject(ctx echo.Context, organization string) error {
	org, err := s.db.FindOrganizationByName(ctx.Request().Context(), organization)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	var dto api.CreateProjectJSONBody
	err = ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	proj, err := s.db.CreateProject(ctx.Request().Context(), org.ID, dto.Name, stringFromPtr(dto.Description))
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusCreated, projToDTO(proj))
}

// (DELETE /v1/organizations/{organization}/project/{name})
func (s *Server) DeleteProject(ctx echo.Context, organization, name string) error {
	proj, err := s.db.FindProjectByName(ctx.Request().Context(), organization, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.db.DeleteProject(ctx.Request().Context(), proj.ID)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.NoContent(http.StatusOK)
}

// (PUT /v1/organizations/{organization}/project/{name})
func (s *Server) UpdateProject(ctx echo.Context, organization, name string) error {
	var dto api.UpdateProjectJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	proj, err := s.db.FindProjectByName(ctx.Request().Context(), organization, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	proj, err = s.db.UpdateProject(ctx.Request().Context(), proj.ID, stringFromPtr(dto.Description))
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, projToDTO(proj))
}

func projToDTO(p *database.Project) *api.Project {
	return &api.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: stringToPtr(p.Description),
		CreatedOn:   types.Date{Time: p.CreatedOn},
		UpdatedOn:   types.Date{Time: p.CreatedOn},
	}
}
