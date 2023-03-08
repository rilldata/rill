package github

import (
	"github.com/google/go-github/v50/github"
)

type Server struct {
	githubClient *github.Client
}
