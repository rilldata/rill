package ai

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/parser"
)

const (
	SearchExternalToolsName = "search_external_tools"
	// defaultExternalSearchResults is the number of tools returned for a query when the LLM does not ask for a number.
	defaultExternalSearchResults = 5
	// maxExternalSearchResults caps the number of tools returned for a query.
	maxExternalSearchResults = 20
	// maxLoadedExternalTools caps the external tools loaded in a session, leaving room for the built-in tools under maxTools.
	maxLoadedExternalTools = 100
)

// SearchExternalTools finds tools of the connected external MCP servers and loads them,
// so their full definitions are offered to the LLM from its next iteration on.
// External tools are loaded on demand to keep their schemas out of requests that don't need them,
// and to support servers with more tools than the LLM providers accept in one request.
type SearchExternalTools struct {
	Runtime *runtime.Runtime
}

var _ Tool[*SearchExternalToolsArgs, *SearchExternalToolsResult] = (*SearchExternalTools)(nil)

type SearchExternalToolsArgs struct {
	Names      []string `json:"names,omitempty" jsonschema:"Exact names of tools to load, as listed in the system prompt."`
	Query      string   `json:"query,omitempty" jsonschema:"Keywords describing the tools to find, such as 'create issue'."`
	MaxResults int      `json:"max_results,omitempty" jsonschema:"Maximum number of tools to return for the query. Defaults to 5."`
}

type SearchExternalToolsResult struct {
	Tools    []*ExternalToolInfo `json:"tools"`
	NotFound []string            `json:"not_found,omitempty"`
	Message  string              `json:"message,omitempty"`
}

// ExternalToolInfo describes a loaded external tool. Its input schema is not included: the LLM gets it as a tool definition.
type ExternalToolInfo struct {
	Name        string `json:"name"`
	Server      string `json:"server"`
	Description string `json:"description,omitempty"`
}

func (t *SearchExternalTools) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        SearchExternalToolsName,
		Title:       "Search external tools",
		Description: "Finds and loads tools of the connected external MCP servers. Pass exact tool names from the system prompt, or keywords describing what you need. The returned tools can be called directly from your next step on, and stay available for the rest of the conversation.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Searching external tools...",
			"openai/toolInvocation/invoked":  "Searched external tools",
		},
	}
}

// CheckAccess only allows the tool inside an agent with external tools.
// This also keeps it off Rill's own MCP server, which never exposes external tools.
func (t *SearchExternalTools) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}
	agent := externalToolsAgent(s)
	return agent != "" && len(s.ExternalToolNames(ctx, agent)) > 0, nil
}

func (t *SearchExternalTools) Handler(ctx context.Context, args *SearchExternalToolsArgs) (*SearchExternalToolsResult, error) {
	if len(args.Names) == 0 && strings.TrimSpace(args.Query) == "" {
		return nil, errors.New("provide either names or query")
	}
	s := GetSession(ctx)
	agent := externalToolsAgent(s)

	var docs []*externalSearchDoc
	for _, srv := range s.ExternalServers(ctx) {
		if !srv.Connected || !srv.AllowsAgent(agent) {
			continue
		}
		for _, mt := range srv.Tools {
			docs = append(docs, &externalSearchDoc{
				info: &ExternalToolInfo{
					Name:        ExternalToolName(srv.Connector, mt.Name),
					Server:      srv.DisplayName,
					Description: sanitizeExternalDescription(mt.Description),
				},
				name:        searchTokens(mt.Name + " " + mt.Title),
				server:      searchTokens(srv.DisplayName + " " + srv.Connector),
				description: searchTokens(mt.Description),
			})
		}
	}

	res := &SearchExternalToolsResult{Tools: []*ExternalToolInfo{}}
	for _, name := range args.Names {
		idx := slices.IndexFunc(docs, func(d *externalSearchDoc) bool { return d.info.Name == name })
		if idx == -1 {
			res.NotFound = append(res.NotFound, name)
			continue
		}
		if !slices.Contains(res.Tools, docs[idx].info) {
			res.Tools = append(res.Tools, docs[idx].info)
		}
	}
	if strings.TrimSpace(args.Query) != "" {
		limit := args.MaxResults
		if limit <= 0 {
			limit = defaultExternalSearchResults
		}
		limit = min(limit, maxExternalSearchResults)
		for _, info := range searchExternalTools(docs, args.Query, limit) {
			if !slices.Contains(res.Tools, info) {
				res.Tools = append(res.Tools, info)
			}
		}
	}
	if len(res.Tools) == 0 {
		res.Message = "No tools matched. Try other keywords, or pass exact names from the system prompt."
		return res, nil
	}

	// The cap counts tools that are currently unavailable too: they are offered again if their server comes back.
	loaded := loadedExternalToolNames(s)
	for _, info := range res.Tools {
		if !slices.Contains(loaded, info.Name) {
			loaded = append(loaded, info.Name)
		}
	}
	if len(loaded) > maxLoadedExternalTools {
		return nil, fmt.Errorf("loading these tools would exceed the limit of %d loaded external tools; continue with the tools already loaded", maxLoadedExternalTools)
	}
	return res, nil
}

// externalToolsAgent returns the agent name used to scope external servers ("analyst" or "developer") for the agent the current call is nested in.
// It returns an empty string outside an agent.
func externalToolsAgent(s *Session) string {
	switch s.agentName() {
	case AnalystAgentName:
		return parser.SkillAgentAnalyst
	case DeveloperAgentName:
		return parser.SkillAgentDeveloper
	default:
		return ""
	}
}

// loadedExternalTools returns the names of the external tools loaded in the session that the given agent may still use, sorted for stable requests.
func loadedExternalTools(ctx context.Context, s *Session, agent string) []string {
	available := s.ExternalToolNames(ctx, agent)
	return slices.DeleteFunc(loadedExternalToolNames(s), func(name string) bool { return !slices.Contains(available, name) })
}

// loadedExternalToolNames returns the names of all external tools loaded in the session, whether or not they are still available, sorted.
func loadedExternalToolNames(s *Session) []string {
	var res []string
	for _, call := range s.Messages(FilterByType(MessageTypeCall), FilterByTool(SearchExternalToolsName)) {
		result, ok := s.Message(FilterByParent(call.ID), FilterByType(MessageTypeResult))
		if !ok || result.ContentType == MessageContentTypeError {
			continue
		}
		content, err := s.UnmarshalMessageContent(result)
		if err != nil {
			continue
		}
		r, ok := content.(*SearchExternalToolsResult)
		if !ok {
			continue
		}
		for _, info := range r.Tools {
			res = append(res, info.Name)
		}
	}
	slices.Sort(res)
	return slices.Compact(res)
}

// externalSearchDoc holds the tokenized fields of an external tool for keyword search.
type externalSearchDoc struct {
	info        *ExternalToolInfo
	name        []string // tool name and title
	server      []string // server display name and connector name
	description []string
}

// searchExternalTools ranks tools by the keywords in query and returns the best matches, up to limit.
// Each query word scores the best field it matches: name or title 3, server 2, description 1.
// A query word matches a tool word that starts with it, so "request" matches "requests".
func searchExternalTools(docs []*externalSearchDoc, query string, limit int) []*ExternalToolInfo {
	words := searchTokens(query)
	slices.Sort(words)
	words = slices.Compact(words)

	type match struct {
		info  *ExternalToolInfo
		score int
	}
	var matches []match
	for _, d := range docs {
		var score int
		for _, w := range words {
			switch {
			case hasTokenWithPrefix(d.name, w):
				score += 3
			case hasTokenWithPrefix(d.server, w):
				score += 2
			case hasTokenWithPrefix(d.description, w):
				score++
			}
		}
		if score > 0 {
			matches = append(matches, match{info: d.info, score: score})
		}
	}
	slices.SortFunc(matches, func(a, b match) int {
		return cmp.Or(b.score-a.score, strings.Compare(a.info.Name, b.info.Name))
	})

	res := make([]*ExternalToolInfo, 0, min(limit, len(matches)))
	for _, m := range matches[:min(limit, len(matches))] {
		res = append(res, m.info)
	}
	return res
}

// searchTokens splits text into lowercase words at non-alphanumeric characters and camelCase boundaries,
// so "listPullRequests" and "list_pull_requests" both yield "list", "pull", "requests".
// A camelCase word is also kept whole, so "GitHub" matches both "github" and "hub".
func searchTokens(text string) []string {
	var res []string
	var word, part strings.Builder
	var parts int
	flush := func() {
		if part.Len() > 0 {
			res = append(res, part.String())
			part.Reset()
			parts++
		}
	}
	var prev rune
	for _, r := range text + " " {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			if parts > 1 {
				res = append(res, word.String())
			}
			word.Reset()
			parts = 0
		} else {
			if unicode.IsUpper(r) && unicode.IsLower(prev) {
				flush()
			}
			part.WriteRune(unicode.ToLower(r))
			word.WriteRune(unicode.ToLower(r))
		}
		prev = r
	}
	return res
}

// hasTokenWithPrefix reports whether any of the tokens starts with prefix.
func hasTokenWithPrefix(tokens []string, prefix string) bool {
	return slices.ContainsFunc(tokens, func(t string) bool { return strings.HasPrefix(t, prefix) })
}
