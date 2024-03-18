package rillv1beta

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"gopkg.in/yaml.v2"
)

const Version = "rill-beta"

type Codec struct {
	Repo       drivers.RepoStore
	InstanceID string
}

func New(repo drivers.RepoStore, instanceID string) *Codec {
	return &Codec{Repo: repo, InstanceID: instanceID}
}

func (c *Codec) IsInit(ctx context.Context) bool {
	_, err := c.Repo.Get(ctx, "rill.yaml")
	return err == nil
}

func (c *Codec) InitEmpty(ctx context.Context, title string) error {
	mockUsersInfo := "# These are example mock users to test your security policies.\n# For more information, see the documentation: https://docs.rilldata.com/manage/security"
	mockUsers := "mock_users:\n- email: john@yourcompany.com\n- email: jane@partnercompany.com"
	err := c.Repo.Put(ctx, "rill.yaml", strings.NewReader(fmt.Sprintf("compiler: %s\n\ntitle: %q\n\n%s\n\n%s", Version, title, mockUsersInfo, mockUsers)))
	if err != nil {
		return err
	}

	gitignore, _ := c.Repo.Get(ctx, ".gitignore")
	if gitignore != "" {
		gitignore += "\n"
	}
	gitignore += ".DS_Store\n\n# Rill\ntmp\n"

	err = c.Repo.Put(ctx, ".gitignore", strings.NewReader(gitignore))
	if err != nil {
		return err
	}

	err = c.Repo.Put(ctx, "sources/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = c.Repo.Put(ctx, "models/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	err = c.Repo.Put(ctx, "dashboards/.gitkeep", strings.NewReader(""))
	if err != nil {
		return err
	}

	return nil
}

func (c *Codec) DeleteSource(ctx context.Context, name string) (string, error) {
	p := path.Join("sources", name+".yaml")
	err := c.Repo.Delete(ctx, p)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (c *Codec) ProjectConfig(ctx context.Context) (*ProjectConfig, error) {
	content, err := c.Repo.Get(ctx, "rill.yaml")
	// rill.yaml is not guaranteed to exist in case of older projects
	if os.IsNotExist(err) {
		return &ProjectConfig{Variables: make(map[string]string)}, nil
	}

	if err != nil {
		return nil, err
	}

	r := &ProjectConfig{Variables: make(map[string]string)}
	if err := yaml.Unmarshal([]byte(content), r); err != nil {
		return nil, err
	}

	return r, nil
}

func HasRillProject(dir string) bool {
	_, err := os.Open(filepath.Join(dir, "rill.yaml"))
	return err == nil
}

func ParseProjectConfig(content []byte) (*ProjectConfig, error) {
	c := &ProjectConfig{Variables: make(map[string]string)}
	if err := yaml.Unmarshal(content, c); err != nil {
		return nil, err
	}

	return c, nil
}
