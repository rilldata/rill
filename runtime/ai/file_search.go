package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const SearchFilesName = "search_files"

type SearchFiles struct {
	Runtime *runtime.Runtime
}

var _ Tool[*SearchFilesArgs, *SearchFilesResult] = (*SearchFiles)(nil)

type SearchFilesArgs struct {
	Pattern       string `json:"pattern" jsonschema:"The pattern to search for. Supports regular expressions."`
	CaseSensitive bool   `json:"case_sensitive,omitempty" jsonschema:"Whether the search should be case-sensitive. Defaults to false."`
	GlobPattern   string `json:"glob_pattern,omitempty" jsonschema:"Optional glob pattern to filter files (e.g., '**/*.sql', 'models/**/*'). If not provided, searches all files."`
}

type SearchFilesResult struct {
	Matches []SearchMatch `json:"matches"`
}

type SearchMatch struct {
	Path     string   `json:"path"`
	Lines    []int    `json:"lines"`
	Snippets []string `json:"snippets"`
}

func (t *SearchFiles) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        SearchFilesName,
		Title:       "Search files",
		Description: "Searches for a pattern across files in the Rill project. Returns matching file paths, line numbers, and snippets. Use this before read_file to discover which files contain specific content.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Searching files...",
			"openai/toolInvocation/invoked":  "Searched files",
		},
	}
}

func (t *SearchFiles) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *SearchFiles) Handler(ctx context.Context, args *SearchFilesArgs) (*SearchFilesResult, error) {
	s := GetSession(ctx)

	// Compile the search pattern
	pattern := args.Pattern
	if !args.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	// Get the glob pattern for filtering files
	globPattern := "**"
	if args.GlobPattern != "" {
		globPattern = args.GlobPattern
	}

	// List all files
	files, err := t.Runtime.ListFiles(ctx, s.InstanceID(), globPattern)
	if err != nil {
		return nil, err
	}

	var matches []SearchMatch
	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Read file contents
		blob, _, err := t.Runtime.GetFile(ctx, s.InstanceID(), file.Path)
		if err != nil {
			// Exit early for cancellations or timeouts
			if errors.Is(err, ctx.Err()) {
				return nil, err
			}
			// Skip files that can't be read
			continue
		}

		// Skip binary or very large files
		if len(blob) > 1024*1024 { // 1MB limit
			continue
		}

		// Search for matches in the file
		lines := strings.Split(blob, "\n")
		var matchingLines []int
		var snippets []string

		for i, line := range lines {
			if !re.MatchString(line) {
				continue
			}
			lineNum := i + 1
			matchingLines = append(matchingLines, lineNum)

			// Create a snippet with context (2 lines before and after)
			start := i - 2
			if start < 0 {
				start = 0
			}
			end := i + 3
			if end > len(lines) {
				end = len(lines)
			}

			snippetLines := []string{}
			for j := start; j < end; j++ {
				prefix := "  "
				if j == i {
					prefix = "> " // Highlight the matching line
				}
				snippetLines = append(snippetLines, fmt.Sprintf("%s%d: %s", prefix, j+1, lines[j]))
			}
			snippets = append(snippets, strings.Join(snippetLines, "\n"))

			// Limit to 5 matches per file to avoid overwhelming results
			if len(matchingLines) >= 5 {
				break
			}
		}

		if len(matchingLines) > 0 {
			matches = append(matches, SearchMatch{
				Path:     file.Path,
				Lines:    matchingLines,
				Snippets: snippets,
			})
		}

		// Limit to 20 matching files to avoid overwhelming results
		if len(matches) >= 20 {
			break
		}
	}

	return &SearchFilesResult{
		Matches: matches,
	}, nil
}
