package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/project"
)

// (GET /v1/organizations/{organization}/projects)
func (s *Server) FindProjects(ctx echo.Context, orgName string) error {
	organization, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	projects, err := organization.QueryProjects().All(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed querying projects: %w", err)
	}

	dtos := make([]*api.Project, len(projects))
	for i, project := range projects {
		dtos[i] = projToDTO(project)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (GET /v1/organizations/{organization}/project/{name})
func (s *Server) FindProject(ctx echo.Context, orgName string, name string) error {
	organization, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	project, err := organization.QueryProjects().Where(project.NameEQ(name)).Only(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, projToDTO(project))
}

// (POST /v1/organizations/{organization}/projects)
func (s *Server) CreateProject(ctx echo.Context, orgName string) error {
	var dto api.CreateProjectJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	// Get the org from OrgName
	organization, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	project, err := s.client.Project.
		Create().
		SetName(dto.Name).
		SetDescription(*dto.Description).
		SetOrganization(organization).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed creating Project: %w", err)
	}

	return ctx.JSON(http.StatusCreated, projToDTO(project))
}

// (DELETE /v1/organizations/{organization}/project/{name})
func (s *Server) DeleteProject(ctx echo.Context, orgName string, name string) error {
	project, err := QueryProjectByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.client.Project.DeleteOne(project).Exec(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, projToDTO(project))
}

// (PUT /v1/organizations/{organization}/project/{name})
func (s *Server) UpdateProject(ctx echo.Context, orgName string, name string) error {
	var dto api.UpdateProjectJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	project, err := QueryProjectByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	// Get the org from OrgName
	org, err := QueryOrganizationByName(ctx.Request().Context(), s.client, orgName)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	projectNew, err := s.client.Project.
		UpdateOne(project).
		// RemoveUsers(). //Do we need to removed edges in case of update?
		// SetName(*dto.Name). // Can't set or udpate as its unique and immutable
		SetDescription(*dto.Description).
		SetOrganization(org).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed updating User: %w", err)
	}

	return ctx.JSON(http.StatusOK, projToDTO(projectNew))
}

func projToDTO(p *ent.Project) *api.Project {
	return &api.Project{
		Id:          fmt.Sprint(p.ID),
		Name:        p.Name,
		Description: stringToPtr(p.Description),
		CreatedOn:   types.Date{Time: p.CreatedOn},
		UpdatedOn:   types.Date{Time: p.CreatedOn},
	}
}

func QueryProjects(ctx context.Context, client *ent.Client) ([]*ent.Project, error) {
	p, err := client.Project.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Projects: %w", err)
	}

	return p, nil
}

func QueryProjectByName(ctx context.Context, client *ent.Client, name string) (*ent.Project, error) {
	p, err := client.Project.
		Query().
		Where(project.NameEQ(name)).
		// `Only` fails if no project found,
		// or more than 1 project returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying User: %w", err)
	}

	return p, nil
}
