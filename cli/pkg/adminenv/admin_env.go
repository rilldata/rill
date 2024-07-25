package adminenv

import (
	"fmt"
)

var EnvURLs = map[string]string{
	"prod":  "https://admin.rilldata.com",
	"stage": "https://admin.rilldata.io",
	"test":  "https://admin.rilldata.in",
	"dev":   "http://localhost:9090",
}

func Infer(adminURL string) (string, error) {
	for env, url := range EnvURLs {
		if url == adminURL {
			return env, nil
		}
	}
	return "", fmt.Errorf("could not infer env from admin URL %q", adminURL)
}

func AdminURL(env string) string {
	u, ok := EnvURLs[env]
	if !ok {
		panic(fmt.Errorf("invalid environment %q", env))
	}
	return u
}
