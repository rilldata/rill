package nameseeds

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

const (
	nameMinLength = 3
	nameMaxLength = 50
)

var (
	nonAlphanumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
	dashesRegex          = regexp.MustCompile("-+")
)

// ForUser generates a list of name seeds from an email and full name.
func ForUser(email, fullName string) []string {
	// Gather candidates
	var seeds []string

	// Based on email
	name, ok := finalizeSeed(strings.Split(email, "@")[0])
	if ok {
		seeds = append(seeds, name)
	}

	// Based on name
	name, ok = finalizeSeed(fullName)
	if ok {
		seeds = append(seeds, name)
	}

	// Find shortest candidate
	var shortest string
	for _, s := range seeds {
		if shortest == "" || len(shortest) > len(s) {
			shortest = s
		}
	}

	// Fallback to UUID-derived
	name, ok = finalizeSeed(shortest + uuid.NewString()[0:8])
	if ok {
		seeds = append(seeds, name)
	}

	return seeds
}

func finalizeSeed(seed string) (string, bool) {
	// replace non-alphanum characters with dashes
	seed = nonAlphanumericRegex.ReplaceAllString(seed, "-")

	// replace repeated dashes with single dashes
	seed = dashesRegex.ReplaceAllString(seed, "-")

	// replace leading dash with underscore
	if seed != "" && seed[0] == '-' {
		seed = "_" + seed[1:]
	}

	// add leading underscore if natural
	if len(seed) == (nameMinLength-1) && seed[0] != '_' {
		seed = "_" + seed
	}

	// add leter prefix is starts with number
	if seed != "" && unicode.IsDigit(rune(seed[0])) {
		seed = "r" + seed
	}

	// cut length if too long
	if len(seed) > nameMaxLength {
		seed = seed[0:nameMaxLength]
	}

	if len(seed) < nameMinLength {
		return "", false
	}

	return strings.ToLower(seed), true
}
