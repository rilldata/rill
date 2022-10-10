package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/user"
)

// (GET /v1/users)
func (s *Server) FindUsers(ctx echo.Context) error {
	// We can check in every Route based on the assest access required
	// other option is to make it part of middleware

	// username, err := GetUserID(ctx)
	// if err != nil {
	// 	return sendError(ctx, http.StatusUnauthorized, "User not found")
	// }

	// action := ctx.Request().Method
	// asset := "perm1"
	// ok := s.authorizer.HasPermission(username, action, asset)
	// if !ok {
	// 	return sendError(ctx, http.StatusUnauthorized, "No permission for this action")
	// }

	users, err := QueryUsers(ctx.Request().Context(), s.client)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.User, len(users))
	for i, user := range users {
		dtos[i] = userToDTO(user)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/users)
func (s *Server) CreateUser(ctx echo.Context) error {
	var dto api.CreateUserJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := s.client.User.
		Create().
		SetName(dto.Name).
		SetUserName(dto.UserName).
		SetDescription(*dto.Description).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed creating User: %w", err)
	}

	return ctx.JSON(http.StatusCreated, userToDTO(user))
}

// (DELETE /v1/user/{name})
func (s *Server) DeleteUser(ctx echo.Context, name string) error {
	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	err = s.client.User.DeleteOne(user).Exec(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, userToDTO(user))
}

// (GET /v1/user/{name})
func (s *Server) FindUser(ctx echo.Context, name string) error {
	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return fmt.Errorf("failed querying User: %w", err)
	}

	return ctx.JSON(http.StatusOK, userToDTO(user))
}

// (PUT /v1/user/{name})
func (s *Server) UpdateUser(ctx echo.Context, name string) error {
	var dto api.UpdateUserJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return fmt.Errorf("failed querying User: %w", err)
	}

	userNew, err := s.client.User.
		UpdateOne(user).
		// RemoveUsers(). //Do we need to removed edges in case of update?
		SetDescription(*dto.Description).
		// AddOrganization(org).
		Save(ctx.Request().Context())
	if err != nil {
		return fmt.Errorf("failed updating User: %w", err)
	}

	return ctx.JSON(http.StatusOK, userToDTO(userNew))
}

// (GET /v1/users/{name}/roles)
func (s *Server) ListRoles(ctx echo.Context, name string) error {
	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	roles, err := user.QueryRole().All(ctx.Request().Context())
	dtos := make([]*api.Role, len(roles))
	for i, role := range roles {
		dtos[i] = roleToDTO(role)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (POST /v1/users/{name}/roles)
func (s *Server) AssignRole(ctx echo.Context, name string) error {
	var dto api.AssignRoleJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	role, err := QueryRoleByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newUser, err := user.Update().AddRole(role).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	roles, err := newUser.QueryRole().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Role, len(roles))
	for i, role := range roles {
		dtos[i] = roleToDTO(role)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

// (PUT /v1/users/{name}/roles)
func (s *Server) RemoveRole(ctx echo.Context, name string) error {
	var dto api.AssignRoleJSONBody
	err := ctx.Bind(&dto)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, "invalid body format")
	}

	user, err := QueryUserByName(ctx.Request().Context(), s.client, name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	role, err := QueryRoleByName(ctx.Request().Context(), s.client, dto.Name)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newUser, err := user.Update().RemoveRole(role).Save(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	roles, err := newUser.QueryRole().All(ctx.Request().Context())
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	dtos := make([]*api.Role, len(roles))
	for i, role := range roles {
		dtos[i] = roleToDTO(role)
	}

	return ctx.JSON(http.StatusOK, dtos)
}

func userToDTO(o *ent.User) *api.User {
	return &api.User{
		Id:          fmt.Sprint(o.ID),
		Description: stringToPtr(o.Description),
		Name:        o.Name,
		UserName:    o.UserName,
		CreatedOn:   types.Date{Time: o.CreatedOn},
		UpdatedOn:   types.Date{Time: o.CreatedOn},
	}
}

func QueryUsers(ctx context.Context, client *ent.Client) ([]*ent.User, error) {
	user, err := client.User.
		Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying User: %w", err)
	}

	return user, nil
}

func QueryUserByName(ctx context.Context, client *ent.Client, name string) (*ent.User, error) {
	user, err := client.User.
		Query().
		Where(user.NameEQ(name)).
		// `Only` fails if no user found,
		// or more than 1 user returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying User: %w", err)
	}

	return user, nil
}

type User struct {
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
}

type Users map[string]User

func LoadUsers(ctx context.Context, client *ent.Client) (Users, error) {
	userList := []User{}
	entUsers, err := QueryUsers(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, user := range entUsers {
		roles := user.QueryRole().AllX(ctx)
		var rolesList []string
		for _, role := range roles {
			rolesList = append(rolesList, role.Name)

		}
		userList = append(userList, User{ID: user.UserName, Roles: rolesList})

	}
	users := Users{}
	for _, user := range userList {
		users[user.ID] = user
	}

	return users, nil
}
