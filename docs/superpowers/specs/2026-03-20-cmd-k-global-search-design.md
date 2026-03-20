# Cmd+K Global Search & Navigation

## Problem

Rill Cloud users navigating between projects, dashboards, reports, and alerts rely on manual click-through navigation. There is no fast, keyboard-driven way to jump to a resource — especially from the org home page where there is no project context. Power users and casual users alike lose time drilling through the sidebar and tab hierarchy.

## Solution

A Cmd+K (Ctrl+K on Windows/Linux) command palette that lets users search across all accessible projects, dashboards, reports, and alerts from anywhere in the app. Results are grouped by type, use Rill's existing resource icons, and navigate on selection.

## Scope

### In scope (v1)
- Global keyboard shortcut: Cmd+K (Mac) / Ctrl+K (Windows/Linux) to open/close
- Searchable entities: Projects, Dashboards (Explore + Canvas), Reports, Alerts
- Results grouped by type: Projects → Dashboards → Reports → Alerts
- Eager prefetch on app load for instant search (no spinners on keystroke)
- Client-side string matching against prefetched data
- Keyboard navigation: arrow keys to move, Enter to open, Esc to close, Cmd+K to toggle
- SvelteKit `goto()` navigation on result selection

### Out of scope (future)
- Organization switching
- Actions/commands (e.g., "create project", "invite user")
- Recently visited / favorites (default view before typing)
- Server-side search endpoint for cross-project resource search at scale
- Visible search bar / UI trigger (keyboard-only for v1)
- Search within dashboard content (dimensions, measures, etc.)

## Design

### UI

- **Trigger:** No visible UI element. Cmd+K / Ctrl+K opens a centered modal overlay.
- **Modal:** Dark overlay backdrop, centered palette (max-width ~520px), positioned toward top of viewport.
- **Search input:** Text input with search icon and placeholder "Search projects, dashboards, reports..."
- **Results:** Grouped by type with uppercase section headers. Each result shows:
  - Resource icon (Folders for projects, ExploreIcon/CanvasIcon for dashboards, ReportIcon, AlertIcon)
  - Resource name
  - Parent project name as breadcrumb context (for dashboards, reports, alerts)
- **Footer:** `↑↓ navigate`, `↵ open`, `⌘K open / close menu`
- **Empty state:** "No results found" message when query matches nothing.
- **Loading state:** Brief spinner shown only if cache isn't warmed yet when palette opens.

### Architecture

#### Component structure

Uses the existing `@rilldata/web-common/components/command` wrappers (which re-export `cmdk-sv` primitives) for consistency with the rest of the codebase.

```
web-admin/src/routes/[organization]/+layout.svelte (org-level)
  └── CommandPalette.svelte
        ├── CommandDialog (shouldFilter={false})
        │     ├── CommandInput
        │     ├── CommandList
        │     │     ├── CommandGroup (Projects)
        │     │     ├── CommandGroup (Dashboards)
        │     │     ├── CommandGroup (Reports)
        │     │     └── CommandGroup (Alerts)
        │     └── CommandEmpty
        └── search-orchestrator.ts
```

Note: The palette is mounted at the `[organization]` layout level (not the root layout) because the org name is required for API calls. This means the palette is available on all org and project pages but not on auth/embed pages — which is the desired behavior.

#### Data flow

`ListResources` is a runtime API, not an admin API. Each project has its own runtime deployment with a separate host, instance ID, and JWT. Fetching resources requires a two-hop authentication:

1. `GetProject` → returns `deployment.runtimeHost`, `deployment.runtimeInstanceId`, and a JWT
2. `ListResources` against that project's runtime host using those credentials

```
App load (org-level layout)
  ├── ListProjectsForOrganizationAndUser(org, pageSize=50) → project list
  │     (userId is optional; omitting it returns the current user's projects)
  │     → immediately build project SearchableItems
  └── for each project (concurrently, batched in groups of 5):
        ├── GetProject(org, project) → runtimeHost, instanceId, JWT
        └── ListResources(instanceId) on runtimeHost → all resources
              → extract Explore, Canvas, Report, Alert resources
              → append to SearchIndex: Array<SearchableItem>

User types in palette
  └── client-side filter on SearchIndex
        └── case-insensitive prefix/contains match on resource name
              └── group by type → render
```

The search index is progressively populated as project credentials resolve. Users can search immediately against whatever is cached so far; results get richer as more projects load.

#### Key types

```typescript
interface SearchableItem {
  name: string;
  type: "project" | "explore" | "canvas" | "report" | "alert";
  projectName: string;
  orgName: string;
  route: string;  // pre-computed navigation URL
}
```

#### Search orchestrator

- `buildSearchIndex()`: Called after prefetch queries resolve. Flattens all projects and their resources into a `SearchableItem[]` array.
- `search(query: string)`: Case-insensitive prefix/contains match against `name` field. Returns results grouped by type, max 5 results per group. Minimum 2 characters to search; below that threshold the palette shows a hint ("Type to search...").
- No debounce needed — client-side filtering is synchronous and fast.

#### Prefetch strategy

- **When:** On org layout mount, triggered from `[organization]/+layout.svelte`. The org name is available from route params; `userId` is omitted (the API defaults to the authenticated user).
- **How:** TanStack Query `prefetchQuery` for project list (with `pageSize: 50`), then for each project: `GetProject` (to obtain runtime credentials) followed by `ListResources` against the project's runtime host. Projects are fetched concurrently in batches of 5 to avoid overwhelming the server. If the project list is paginated (`nextPageToken` present), fetch subsequent pages in the background.
- **Cache:** `staleTime: 5 minutes`. TanStack Query manages cache lifecycle and background refetching. For runtime JWTs: the `staleTime` on `GetProject` queries should be set shorter than the JWT TTL so that cached credentials are refreshed before expiry. When a `ListResources` call fails with an auth error, the orchestrator should invalidate the corresponding `GetProject` cache entry and retry once.
- **Progressive loading:** The search index is usable as soon as the project list loads (project-level search works immediately). Resource-level results appear progressively as each project's runtime responds. The palette does not block on all projects loading.
- **Scale concern:** For users with many projects (>20), only prefetch the first 20 projects on layout mount. A subtle indicator in the palette ("Searching 20 of 45 projects") communicates partial coverage. The remaining projects load in the background; the indicator updates as more complete.

#### Keyboard handling

- Global `keydown` listener on `window` in `[organization]/+layout.svelte`.
- Detect platform: use existing pattern from codebase (`navigator.userAgent.includes("Macintosh")`), extracted to a shared utility if not already available.
- Mac: `event.metaKey && event.key === "k"` → toggle palette open/close.
- Windows/Linux: `event.ctrlKey && event.key === "k"` → toggle palette open/close.
- `event.preventDefault()` to avoid browser default behavior (Ctrl+K focuses address bar in some browsers).
- Esc closes the palette (handled by `cmdk-sv` dialog natively).
- The keyboard listener is not active on embed or public pages (the palette is mounted at the org layout level, which excludes these routes).

#### Navigation

On result selection:
- Close the palette
- Call SvelteKit `goto(item.route)` with the pre-computed URL

Route patterns:
- Project: `/{orgName}/{projectName}`
- Explore dashboard: `/{orgName}/{projectName}/explore/{dashboardName}`
- Canvas dashboard: `/{orgName}/{projectName}/canvas/{dashboardName}`
- Report: `/{orgName}/{projectName}/-/reports/{reportName}` (uses resource name as route param)
- Alert: `/{orgName}/{projectName}/-/alerts/{alertName}` (uses resource name as route param; verify at implementation time whether name or ID is expected by the `[alert]` route)

### File structure

```
web-admin/src/features/command-palette/
  ├── CommandPalette.svelte        # Main component, mounted in [organization]/+layout.svelte
  ├── CommandPaletteItem.svelte    # Individual result row
  ├── search-orchestrator.ts       # Prefetch, index building, search logic
  ├── types.ts                     # SearchableItem and related types
  └── route-builders.ts            # URL generation per resource type
```

### Dependencies

- `cmdk-sv` — already in `web-common`, provides Command primitives
- Existing Rill icons — `ExploreIcon`, `CanvasIcon`, `ReportIcon`, `AlertIcon` from `web-common/src/components/icons/`
- Lucide `Folders` icon — for project results
- TanStack Query — already used throughout the app
- SvelteKit `goto` — for navigation

No new dependencies required.

## Error handling

- **Prefetch failure:** If a project's `GetProject` or `ListResources` call fails (permissions, deleted project, runtime unavailable), silently skip it. Other projects' results still show.
- **All prefetch fails:** Palette opens with empty state; show "Unable to load search data" instead of "No results."
- **Navigation failure:** If `goto()` fails (resource deleted between search and click), SvelteKit's error handling catches it naturally.

## Testing

- **Unit tests:** Search orchestrator — verify index building, filtering, grouping, edge cases (empty query, special characters, no results).
- **Component tests:** CommandPalette — verify open/close on Cmd+K, keyboard navigation through results, result selection triggers navigation.
- **E2E test:** Open palette → type query → select result → verify navigation to correct page.
