package publicemail

import (
	"bufio"
	"embed"
	"strings"
	"sync"
)

const publicEmailProvidersListFile = "public_email_providers_list"

//go:embed public_email_providers_list
var ef embed.FS

var (
	once      sync.Once
	providers map[string]bool
)

// read public email providers list file line by line and load it into a map
func loadPublicEmailProvidersList() (map[string]bool, error) {
	providers := make(map[string]bool)

	f, err := ef.Open(publicEmailProvidersListFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		domain := scanner.Text()
		if strings.HasPrefix(domain, "#") {
			continue
		}
		providers[strings.TrimSpace(scanner.Text())] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

// IsPublic returns true if the domain is a public email domain
func IsPublic(domain string) bool {
	once.Do(func() {
		var err error
		providers, err = loadPublicEmailProvidersList()
		if err != nil {
			panic(err)
		}
	})

	_, ok := providers[domain]
	return ok
}
