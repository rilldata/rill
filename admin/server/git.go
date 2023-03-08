package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/eventhandler"
	"go.uber.org/zap"
)

// It MAY be possible to make handleEvent a common handler for all originators like github,gitlab etc.
// In this case the validations and parsing should be part of eventhandler.Handler in a separate Parse method.
// The server then can maintain a map of origin vs handlers.
// This should then get the right handler basis path params and run Parse in sync and Process in async.
func (s *Server) handleEvent(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	payload, err := github.ValidatePayload(req, []byte(s.opts.GithubAppWebhookSecret))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := context.Background()

	// TODO :: this should be processed asynchronously since github webhooks have timeouts of 10 seconds
	err = s.handler.Process(ctx, event)
	if err != nil {
		if errors.Is(err, eventhandler.ErrInvalidEvent) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) connectProject(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	// assuming some middleware already checks and redirects user to login page before it reaches here
	values := req.URL.Query()
	orgName := pathParams["organization"]
	remote, err := url.QueryUnescape(values.Get("remote"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	projectName := values.Get("project_name")
	prodBranch := values.Get("prod_branch")

	ctx := req.Context()
	org, err := s.admin.DB.FindOrganizationByName(ctx, orgName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	endpoint, err := transport.NewEndpoint(remote)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// todo :: find a better way to do this
	fullName := parseRepoPath(endpoint.Path, endpoint.Protocol)
	project, err := s.getOrCreate(ctx, org, projectName, remote, fullName, prodBranch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if project.GithubAppInstallID != 0 {
		// we already know installation id
		// should we handle cases when user is trying to add the installation again ??
		w.WriteHeader(http.StatusAlreadyReported)
		return
	}

	owner, repo, found := strings.Cut(fullName, "/")
	if !found {
		// invalid remote
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	installation, response, err := s.githubClient.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			// we are going to receive this state back in callback once user has installed the app
			state := installationState{Project: projectName, Org: orgName}
			encodedState, err := state.encode()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			installLink := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", s.opts.GithubAppName, encodedState)
			http.Redirect(w, req, installLink, http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(response.StatusCode)
		return
	}

	// we already have access
	installationID := installation.GetID()
	if installationID != 0 {
		project.GithubAppInstallID = installation.GetID()
		project.GitURL, err = s.httpRemote(ctx, installationID, project.GitURL, owner, repo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	project, err = s.admin.DB.UpdateProject(ctx, project)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.logger.Debug("updated project ", zap.String("projectId", project.ID))
}

// installSetupCallback gets called once the user has installed the app on the repository
// We leverage this to verify that user installed the app on the repo that we need and navigate user to correct pages
func (s *Server) installSetupCallback(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	ctx := req.Context()
	values := req.URL.Query()
	stateString := values.Get("state")
	if stateString == "" {
		s.logger.Error("not state found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	installationState, err := newInstallationState(stateString)
	if err != nil {
		// redirect to bad request
		s.logger.Error("unable to parse installation state ", zap.Error(err), zap.String("state", stateString))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// verify that we have the project
	project, err := s.admin.DB.FindProjectByName(ctx, installationState.Org, installationState.Project)
	if err != nil {
		// todo :: revert to some page saying project is not connected ???
		s.logger.Error("project fetch from fb failed ", zap.Error(err),
			zap.String("org", installationState.Org),
			zap.String("project", installationState.Project))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// we have already received installation event
	if project.GithubAppInstallID != 0 {
		// todo :: redirect to success page
		w.WriteHeader(http.StatusOK)
		return
	}

	owner, repo, _ := strings.Cut(project.GitFullName, "/")
	// missed/delayed installation event, verify we have access
	installation, response, err := s.githubClient.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			// redirect to failure page ?
			s.logger.Error("app still does not have access to repo ", zap.String("repo", project.GitFullName))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	installationID := installation.GetID()
	project.GithubAppInstallID = installationID

	// once we have access, change git url to use http url instead of ssh url for github app credentials to work
	project.GitURL, err = s.httpRemote(ctx, installationID, project.GitURL, owner, repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ignoring error
	_, _ = s.admin.DB.UpdateProject(ctx, project)
	// todo :: redirect to success page
	w.WriteHeader(http.StatusOK)
}

func (s *Server) httpRemote(ctx context.Context, installationID int64, remote, owner, repoName string) (string, error) {
	client, err := githubInstallationClient(s.opts, installationID)
	if err != nil {
		return "", err
	}

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return "", err
	}

	httpURL := repo.GetCloneURL()
	if httpURL == "" {
		// should hopefully never happen
		return "", fmt.Errorf("no http url")
	}
	httpEndpoint, _ := transport.NewEndpoint(httpURL)
	httpEndpoint.User = "__githubapp_installation_id__"
	httpEndpoint.Password = fmt.Sprint(installationID)
	return httpEndpoint.String(), nil
}

func (s *Server) getOrCreate(ctx context.Context, org *database.Organization, projectName, remote, fullName, prodBranch string) (*database.Project, error) {
	project, err := s.admin.DB.FindProjectByName(ctx, org.Name, projectName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			project := &database.Project{
				OrganizationID: org.ID,
				Name:           projectName,
				Description:    "",
				GitURL:         remote,
				GitFullName:    fullName,
			}
			if prodBranch != "" {
				project.ProductionBranch = prodBranch
			}
			return s.admin.DB.CreateProject(ctx, org.ID, project)
		}
		return nil, err
	}
	return project, err
}

type installationState struct {
	Project string
	Org     string
}

func (i *installationState) encode() (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func newInstallationState(in string) (*installationState, error) {
	dec, err := hex.DecodeString(in)
	if err != nil {
		return nil, err
	}
	installationState := &installationState{}
	err = json.Unmarshal(dec, installationState)
	return installationState, err
}

// converts /owner/repo.git to owner/repo for http
// converts owner/repo.git to owner/repo for ssh
func parseRepoPath(path, protocol string) string {
	if strings.Contains(protocol, "http") {
		_, path, _ = strings.Cut(path, "/")
	}
	path, _, _ = strings.Cut(path, ".git")
	return path
}

// github client that works for specific installation
func githubInstallationClient(conf *Options, installationID int64) (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the app ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.New(tr, conf.GithubAppID, installationID, []byte(conf.GithubAppPrivateKey))
	if err != nil {
		return nil, err
	}

	// Use installation transport with github.com/google/go-github
	return github.NewClient(&http.Client{Transport: itr}), nil
}
