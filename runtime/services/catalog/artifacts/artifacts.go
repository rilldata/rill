package artifacts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

var Artifacts = make(map[string]Artifact)

var (
	ErrFileRead        = errors.New("failed to read artifact")
	ErrInvalidFileName = errors.New("invalid file name")
)

func Register(name string, artifact Artifact) {
	if Artifacts[name] != nil {
		panic(fmt.Errorf("already registered artifact type with name '%s'", name))
	}
	Artifacts[name] = artifact
}

type Artifact interface {
	DeSerialise(ctx context.Context, filePath string, blob string) (*drivers.CatalogEntry, error)
	Serialise(ctx context.Context, catalogObject *drivers.CatalogEntry) (string, error)
}

func Read(ctx context.Context, repoStore drivers.RepoStore, registryStore drivers.RegistryStore, instID, filePath string) (*drivers.CatalogEntry, error) {
	extension := fileutil.FullExt(filePath)
	artifact, ok := Artifacts[extension]
	if !ok {
		return nil, fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := repoStore.Get(ctx, instID, filePath)
	if err != nil {
		return nil, ErrFileRead
	}

	instance, err := registryStore.FindInstance(ctx, instID)
	if err != nil {
		return nil, err
	}

	// this is required in order to be able to use .env.KEY and not .KEY in template placeholders
	env := map[string]map[string]string{"env": instance.InstanceVariables()}

	// Add Sprig template functions (removing functions that leak host info)
	// Derived from Helm: https://github.com/helm/helm/blob/main/pkg/engine/funcs.go
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// convert templatised artifact
	t, err := template.New("source").Funcs(funcMap).Option("missingkey=error").Parse(blob)
	if err != nil {
		return nil, err
	}

	bw := new(bytes.Buffer)
	if err := t.Execute(bw, env); err != nil {
		return nil, err
	}

	catalog, err := artifact.DeSerialise(ctx, filePath, bw.String())
	if err != nil {
		return nil, err
	}

	if !IsValidName(fileutil.Stem(filePath)) {
		return nil, ErrInvalidFileName
	}

	catalog.Path = filePath
	return catalog, nil
}

func Write(ctx context.Context, repoStore drivers.RepoStore, instID string, catalog *drivers.CatalogEntry) error {
	extension := fileutil.FullExt(catalog.Path)
	artifact, ok := Artifacts[extension]
	if !ok {
		return fmt.Errorf("no artifact found for %s", extension)
	}

	blob, err := artifact.Serialise(ctx, catalog)
	if err != nil {
		return err
	}

	return repoStore.Put(ctx, instID, catalog.Path, strings.NewReader(blob))
}

var regex = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func IsValidName(itemName string) bool {
	return regex.MatchString(itemName)
}

var invalidChars = regexp.MustCompile(`[^a-zA-Z_\d]`)

// SanitizedName returns a sanitized name for an artifact from file path.
func SanitizedName(filePath string) string {
	name := invalidChars.ReplaceAllString(fileutil.Stem(filePath), "_")
	if unicode.IsNumber(rune(name[0])) {
		// prepend underscore if name starts with a number
		name = "_" + name
	}
	return name
}
