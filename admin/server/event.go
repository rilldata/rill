package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/server/eventhandler"
)

// It MAY be possible to make handleEvent a common handler for all originators like github,gitlab etc.
// In this case the validations and parsing should be part of eventhandler.Handler in a separate Parse method.
// The server then can maintain a map of origin vs handlers.
// This should then get the right handler basis path params and run Parse in sync and Process in async.
func (s *Server) handleEvent(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	payload, err := github.ValidatePayload(req, []byte(s.conf.GithubSecretKey))
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
