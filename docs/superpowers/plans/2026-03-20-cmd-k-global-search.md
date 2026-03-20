# Cmd+K Global Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Cmd+K command palette to Rill Cloud that lets users search across projects, dashboards, reports, and alerts from anywhere in the org.

**Architecture:** A `CommandPalette` component mounted at the org layout level uses `cmdk-sv` (via web-common wrappers) with `shouldFilter={false}`. A search orchestrator eagerly prefetches project and resource data on layout mount, builds a client-side search index, and filters it synchronously on each keystroke.

**Tech Stack:** Svelte 4, cmdk-sv, TanStack Query, SvelteKit, TypeScript

**Spec:** `docs/superpowers/specs/2026-03-20-cmd-k-global-search-design.md`

---

## File Map

```
NEW FILES:
  web-admin/src/features/command-palette/types.ts              # SearchableItem type + result grouping types
  web-admin/src/features/command-palette/route-builders.ts      # URL generation per resource type
  web-admin/src/features/command-palette/search-orchestrator.ts # Index building, search/filter logic
  web-admin/src/features/command-palette/resource-prefetch.ts   # Two-hop runtime auth + resource fetching per project
  web-admin/src/features/command-palette/CommandPaletteItem.svelte  # Single result row component
  web-admin/src/features/command-palette/CommandPalette.svelte  # Main palette component with cmdk-sv

MODIFIED FILES:
  web-admin/src/routes/[organization]/+layout.svelte            # Mount CommandPalette component

TEST FILES:
  web-admin/src/features/command-palette/search-orchestrator.spec.ts  # Unit tests for orchestrator
  web-admin/src/features/command-palette/route-builders.spec.ts       # Unit tests for route builders
```

---

### Task 1: Types and Route Builders

**Files:**
- Create: `web-admin/src/features/command-palette/types.ts`
- Create: `web-admin/src/features/command-palette/route-builders.ts`
- Create: `web-admin/src/features/command-palette/route-builders.spec.ts`

- [ ] **Step 1: Create the types file**

```typescript
// web-admin/src/features/command-palette/types.ts

export type SearchableItemType =
  | "project"
  | "explore"
  | "canvas"
  | "report"
  | "alert";

export interface SearchableItem {
  name: string;
  type: SearchableItemType;
  projectName: string;
  orgName: string;
  route: string;
}

export interface GroupedResults {
  projects: SearchableItem[];
  dashboards: SearchableItem[];
  reports: SearchableItem[];
  alerts: SearchableItem[];
}
```

- [ ] **Step 2: Write failing tests for route builders**

```typescript
// web-admin/src/features/command-palette/route-builders.spec.ts
import { describe, it, expect } from "vitest";
import { buildRoute } from "./route-builders";

describe("buildRoute", () => {
  it("builds project route", () => {
    expect(buildRoute("project", "acme", "analytics", "analytics")).toBe(
      "/acme/analytics",
    );
  });

  it("builds explore dashboard route", () => {
    expect(
      buildRoute("explore", "acme", "analytics", "revenue-overview"),
    ).toBe("/acme/analytics/explore/revenue-overview");
  });

  it("builds canvas dashboard route", () => {
    expect(
      buildRoute("canvas", "acme", "analytics", "campaign-tracker"),
    ).toBe("/acme/analytics/canvas/campaign-tracker");
  });

  it("builds report route", () => {
    expect(buildRoute("report", "acme", "analytics", "weekly-report")).toBe(
      "/acme/analytics/-/reports/weekly-report",
    );
  });

  it("builds alert route", () => {
    expect(buildRoute("alert", "acme", "analytics", "revenue-drop")).toBe(
      "/acme/analytics/-/alerts/revenue-drop",
    );
  });
});
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `npx vitest run web-admin/src/features/command-palette/route-builders.spec.ts`
Expected: FAIL — module not found

- [ ] **Step 4: Implement route builders**

```typescript
// web-admin/src/features/command-palette/route-builders.ts
import type { SearchableItemType } from "./types";

export function buildRoute(
  type: SearchableItemType,
  orgName: string,
  projectName: string,
  resourceName: string,
): string {
  switch (type) {
    case "project":
      return `/${orgName}/${projectName}`;
    case "explore":
      return `/${orgName}/${projectName}/explore/${resourceName}`;
    case "canvas":
      return `/${orgName}/${projectName}/canvas/${resourceName}`;
    case "report":
      return `/${orgName}/${projectName}/-/reports/${resourceName}`;
    case "alert":
      return `/${orgName}/${projectName}/-/alerts/${resourceName}`;
  }
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `npx vitest run web-admin/src/features/command-palette/route-builders.spec.ts`
Expected: All 5 tests PASS

- [ ] **Step 6: Commit**

```bash
git add web-admin/src/features/command-palette/types.ts web-admin/src/features/command-palette/route-builders.ts web-admin/src/features/command-palette/route-builders.spec.ts
git commit -m "feat(command-palette): add types and route builders"
```

---

### Task 2: Search Orchestrator

**Files:**
- Create: `web-admin/src/features/command-palette/search-orchestrator.ts`
- Create: `web-admin/src/features/command-palette/search-orchestrator.spec.ts`

- [ ] **Step 1: Write failing tests for search orchestrator**

```typescript
// web-admin/src/features/command-palette/search-orchestrator.spec.ts
import { describe, it, expect } from "vitest";
import { searchIndex, groupResults } from "./search-orchestrator";
import type { SearchableItem } from "./types";

const items: SearchableItem[] = [
  {
    name: "acme-analytics",
    type: "project",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics",
  },
  {
    name: "acme-marketing",
    type: "project",
    projectName: "acme-marketing",
    orgName: "acme",
    route: "/acme/acme-marketing",
  },
  {
    name: "Revenue Overview",
    type: "explore",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/explore/revenue-overview",
  },
  {
    name: "Campaign Tracker",
    type: "canvas",
    projectName: "acme-marketing",
    orgName: "acme",
    route: "/acme/acme-marketing/canvas/campaign-tracker",
  },
  {
    name: "Weekly Revenue Report",
    type: "report",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/-/reports/weekly-revenue-report",
  },
  {
    name: "Revenue Drop Alert",
    type: "alert",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/-/alerts/revenue-drop-alert",
  },
];

describe("searchIndex", () => {
  it("returns empty groups for queries shorter than 2 chars", () => {
    const result = searchIndex(items, "a");
    expect(result.projects).toHaveLength(0);
    expect(result.dashboards).toHaveLength(0);
    expect(result.reports).toHaveLength(0);
    expect(result.alerts).toHaveLength(0);
  });

  it("matches projects by name (case-insensitive)", () => {
    const result = searchIndex(items, "acme");
    expect(result.projects).toHaveLength(2);
  });

  it("matches dashboards by name", () => {
    const result = searchIndex(items, "revenue");
    expect(result.dashboards).toHaveLength(1);
    expect(result.dashboards[0].name).toBe("Revenue Overview");
  });

  it("groups explore and canvas under dashboards", () => {
    const result = searchIndex(items, "er"); // matches "Tracker" and "Overview"
    // "Revenue Overview" (explore), "Campaign Tracker" (canvas) both match
    expect(
      result.dashboards.every(
        (d) => d.type === "explore" || d.type === "canvas",
      ),
    ).toBe(true);
  });

  it("matches reports", () => {
    const result = searchIndex(items, "weekly");
    expect(result.reports).toHaveLength(1);
    expect(result.reports[0].name).toBe("Weekly Revenue Report");
  });

  it("matches alerts", () => {
    const result = searchIndex(items, "drop");
    expect(result.alerts).toHaveLength(1);
    expect(result.alerts[0].name).toBe("Revenue Drop Alert");
  });

  it("limits results to 5 per group", () => {
    const manyProjects: SearchableItem[] = Array.from({ length: 10 }, (_, i) => ({
      name: `project-${i}`,
      type: "project" as const,
      projectName: `project-${i}`,
      orgName: "acme",
      route: `/acme/project-${i}`,
    }));
    const result = searchIndex(manyProjects, "project");
    expect(result.projects).toHaveLength(5);
  });

  it("returns empty groups for empty query", () => {
    const result = searchIndex(items, "");
    expect(result.projects).toHaveLength(0);
    expect(result.dashboards).toHaveLength(0);
  });

  it("matches across name and project name", () => {
    const result = searchIndex(items, "marketing");
    expect(result.projects).toHaveLength(1);
    expect(result.dashboards).toHaveLength(1); // Campaign Tracker's projectName contains "marketing"
  });
});

describe("groupResults", () => {
  it("separates items into correct groups", () => {
    const grouped = groupResults(items);
    expect(grouped.projects).toHaveLength(2);
    expect(grouped.dashboards).toHaveLength(2);
    expect(grouped.reports).toHaveLength(1);
    expect(grouped.alerts).toHaveLength(1);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npx vitest run web-admin/src/features/command-palette/search-orchestrator.spec.ts`
Expected: FAIL — module not found

- [ ] **Step 3: Implement search orchestrator**

```typescript
// web-admin/src/features/command-palette/search-orchestrator.ts
import type { SearchableItem, GroupedResults } from "./types";

const MAX_RESULTS_PER_GROUP = 5;
const MIN_QUERY_LENGTH = 2;

export function searchIndex(
  items: SearchableItem[],
  query: string,
): GroupedResults {
  if (query.length < MIN_QUERY_LENGTH) {
    return { projects: [], dashboards: [], reports: [], alerts: [] };
  }

  const q = query.toLowerCase();
  const matched = items.filter(
    (item) =>
      item.name.toLowerCase().includes(q) ||
      item.projectName.toLowerCase().includes(q),
  );

  return groupResults(matched, MAX_RESULTS_PER_GROUP);
}

export function groupResults(
  items: SearchableItem[],
  limit?: number,
): GroupedResults {
  const groups: GroupedResults = {
    projects: [],
    dashboards: [],
    reports: [],
    alerts: [],
  };

  for (const item of items) {
    switch (item.type) {
      case "project":
        groups.projects.push(item);
        break;
      case "explore":
      case "canvas":
        groups.dashboards.push(item);
        break;
      case "report":
        groups.reports.push(item);
        break;
      case "alert":
        groups.alerts.push(item);
        break;
    }
  }

  if (limit) {
    groups.projects = groups.projects.slice(0, limit);
    groups.dashboards = groups.dashboards.slice(0, limit);
    groups.reports = groups.reports.slice(0, limit);
    groups.alerts = groups.alerts.slice(0, limit);
  }

  return groups;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run web-admin/src/features/command-palette/search-orchestrator.spec.ts`
Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
git add web-admin/src/features/command-palette/search-orchestrator.ts web-admin/src/features/command-palette/search-orchestrator.spec.ts
git commit -m "feat(command-palette): add search orchestrator with filtering and grouping"
```

---

### Task 3: CommandPaletteItem Component

**Files:**
- Create: `web-admin/src/features/command-palette/CommandPaletteItem.svelte`

- [ ] **Step 1: Create the result item component**

This component renders a single row in the palette. It shows the resource icon, name, and project breadcrumb.

```svelte
<!-- web-admin/src/features/command-palette/CommandPaletteItem.svelte -->
<script lang="ts">
  import { Folders } from "lucide-svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import type { SearchableItem } from "./types";

  export let item: SearchableItem;

  const iconComponents = {
    project: null, // handled separately (Folders is Lucide, not Rill icon)
    explore: ExploreIcon,
    canvas: CanvasIcon,
    report: ReportIcon,
    alert: AlertIcon,
  };

  $: IconComponent = iconComponents[item.type];
  $: showBreadcrumb = item.type !== "project";
</script>

<div class="flex items-center gap-2.5 w-full">
  <div class="flex-none w-4 h-4 text-gray-400">
    {#if item.type === "project"}
      <Folders size={16} />
    {:else if IconComponent}
      <svelte:component this={IconComponent} size="16px" />
    {/if}
  </div>
  <div class="flex flex-col min-w-0">
    <span class="text-sm text-gray-200 truncate">{item.name}</span>
    {#if showBreadcrumb}
      <span class="text-xs text-gray-500 truncate">{item.projectName}</span>
    {/if}
  </div>
</div>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/command-palette/CommandPaletteItem.svelte
git commit -m "feat(command-palette): add result item component with resource icons"
```

---

### Task 4: CommandPalette Component

**Files:**
- Create: `web-admin/src/features/command-palette/CommandPalette.svelte`

This is the main palette component. It handles:
- Open/close state via Cmd+K
- Search input with cmdk-sv (shouldFilter=false)
- Rendering grouped results
- Navigation on selection

- [ ] **Step 1: Create the CommandPalette component**

```svelte
<!-- web-admin/src/features/command-palette/CommandPalette.svelte -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    Dialog as CommandDialog,
    Input as CommandInput,
    List as CommandList,
    Empty as CommandEmpty,
    Group as CommandGroup,
    Item as CommandItem,
  } from "@rilldata/web-common/components/command";
  import { createAdminServiceListProjectsForOrganizationAndUser } from "@rilldata/web-admin/client";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    createRuntimeServiceListResources,
  } from "@rilldata/web-common/runtime-client/v2";
  import { getCloudRuntimeClient } from "$lib/runtime-client";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { searchIndex } from "./search-orchestrator";
  import { buildRoute } from "./route-builders";
  import CommandPaletteItem from "./CommandPaletteItem.svelte";
  import type { SearchableItem } from "./types";

  let open = false;
  let query = "";

  $: orgName = $page.params.organization;

  // Prefetch project list
  $: projectListQuery = createAdminServiceListProjectsForOrganizationAndUser(
    orgName,
    { pageSize: 50 },
    {
      query: {
        enabled: !!orgName,
        staleTime: 5 * 60 * 1000,
      },
    },
  );

  // Build search index from projects
  // API response has `projects: V1Project[]` (not `projectRoles`)
  $: projectItems = buildProjectItems(orgName, $projectListQuery.data?.projects);

  // Resource items are populated by the prefetch logic in Task 6
  let resourceItems: SearchableItem[] = [];

  $: searchItems = [...projectItems, ...resourceItems];
  $: results = searchIndex(searchItems, query);
  $: hasResults =
    results.projects.length > 0 ||
    results.dashboards.length > 0 ||
    results.reports.length > 0 ||
    results.alerts.length > 0;

  function buildProjectItems(
    org: string,
    projects: Array<{ name?: string }> | undefined,
  ): SearchableItem[] {
    if (!projects) return [];
    return projects
      .filter((p) => p.name)
      .map((p) => ({
        name: p.name!,
        type: "project" as const,
        projectName: p.name!,
        orgName: org,
        route: buildRoute("project", org, p.name!, p.name!),
      }));
  }

  function handleSelect(item: SearchableItem) {
    open = false;
    query = "";
    void goto(item.route);
  }

  function handleKeydown(e: KeyboardEvent) {
    const isMac = window.navigator.userAgent.includes("Macintosh");
    if (e[isMac ? "metaKey" : "ctrlKey"] && e.key === "k") {
      e.preventDefault();
      open = !open;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<CommandDialog bind:open shouldFilter={false}>
  <CommandInput
    placeholder="Search projects, dashboards, reports..."
    bind:value={query}
  />
  <CommandList>
    {#if query.length < 2}
      <div class="py-6 text-center text-sm text-gray-500">
        Type to search...
      </div>
    {:else if !hasResults}
      <CommandEmpty>No results found.</CommandEmpty>
    {:else}
      {#if results.projects.length > 0}
        <CommandGroup heading="Projects">
          {#each results.projects as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.dashboards.length > 0}
        <CommandGroup heading="Dashboards">
          {#each results.dashboards as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.reports.length > 0}
        <CommandGroup heading="Reports">
          {#each results.reports as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.alerts.length > 0}
        <CommandGroup heading="Alerts">
          {#each results.alerts as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}
    {/if}
  </CommandList>

  <div
    class="flex items-center gap-4 px-4 py-2 border-t border-gray-700 text-[11px] text-gray-500"
  >
    <span>↑↓ navigate</span>
    <span>↵ open</span>
    <span class="ml-auto flex items-center gap-1">
      <kbd
        class="bg-gray-800 border border-gray-600 rounded px-1.5 py-0.5 text-[10px] text-gray-400"
      >
        {window.navigator.userAgent.includes("Macintosh") ? "⌘" : "Ctrl+"}K
      </kbd>
      open / close menu
    </span>
  </div>
</CommandDialog>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/command-palette/CommandPalette.svelte
git commit -m "feat(command-palette): add main palette component with project search"
```

---

### Task 5: Mount in Org Layout

**Files:**
- Modify: `web-admin/src/routes/[organization]/+layout.svelte`

- [ ] **Step 1: Add CommandPalette to the org layout**

In `web-admin/src/routes/[organization]/+layout.svelte`, add the import and component:

After the existing imports (around line 5), add:
```typescript
import CommandPalette from "@rilldata/web-admin/features/command-palette/CommandPalette.svelte";
```

After the `<WelcomeToRillCloudDialog>` component (around line 23), add:
```svelte
<CommandPalette />
```

- [ ] **Step 2: Verify the app builds**

Run: `npm run build -w web-admin`
Expected: Build succeeds with no errors

- [ ] **Step 3: Commit**

```bash
git add web-admin/src/routes/\[organization\]/+layout.svelte
git commit -m "feat(command-palette): mount palette in org layout"
```

---

### Task 6: Resource Prefetch (Dashboards, Reports, Alerts)

**Files:**
- Create: `web-admin/src/features/command-palette/resource-prefetch.ts`
- Modify: `web-admin/src/features/command-palette/CommandPalette.svelte`

This task adds the two-hop resource prefetch. The challenge: in Svelte 4 you can't dynamically create N reactive TanStack Query subscriptions in a loop. The solution is to use imperative `queryClient.fetchQuery` calls inside `onMount` / reactive statements, rather than the `createXxx` query hooks.

- [ ] **Step 1: Create the resource prefetch module**

```typescript
// web-admin/src/features/command-palette/resource-prefetch.ts
import {
  type QueryClient,
} from "@tanstack/svelte-query";
import {
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
} from "@rilldata/web-admin/client";
import {
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client/v2";
import { getCloudRuntimeClient } from "$lib/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { buildRoute } from "./route-builders";
import type { SearchableItem } from "./types";

const SEARCHABLE_KINDS = new Set([
  ResourceKind.Explore,
  ResourceKind.Canvas,
  ResourceKind.Report,
  ResourceKind.Alert,
]);

const RESOURCE_KIND_TO_TYPE: Record<string, SearchableItem["type"]> = {
  [ResourceKind.Explore]: "explore",
  [ResourceKind.Canvas]: "canvas",
  [ResourceKind.Report]: "report",
  [ResourceKind.Alert]: "alert",
};

const BATCH_SIZE = 5;
const STALE_TIME = 5 * 60 * 1000;

/**
 * Fetches resources for a single project via the two-hop auth pattern:
 * 1. GetProject → runtime credentials (host, instanceId, JWT)
 * 2. ListResources → all resources in that project's runtime
 *
 * Returns SearchableItem[] for the searchable resource types.
 * Returns empty array on failure (project unavailable, auth error, etc.)
 */
async function fetchProjectResources(
  queryClient: QueryClient,
  orgName: string,
  projectName: string,
): Promise<SearchableItem[]> {
  try {
    // Step 1: Get runtime credentials via GetProject
    const projectData = await queryClient.fetchQuery({
      queryKey: getAdminServiceGetProjectQueryKey(orgName, projectName),
      queryFn: ({ signal }) =>
        adminServiceGetProject(orgName, projectName, undefined, signal),
      staleTime: STALE_TIME,
    });

    const host = projectData.deployment?.runtimeHost;
    const instanceId = projectData.deployment?.runtimeInstanceId;
    const jwt = projectData.jwt;

    if (!host || !instanceId) return [];

    // Step 2: Create runtime client and list resources
    const client = getCloudRuntimeClient({
      host,
      instanceId,
      jwt: jwt ? { token: jwt } : undefined,
    });

    const resourceData = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {}),
      queryFn: ({ signal }) =>
        runtimeServiceListResources(client, {}, { signal }),
      staleTime: STALE_TIME,
    });

    if (!resourceData.resources) return [];

    // Step 3: Filter to searchable kinds and map to SearchableItem
    return resourceData.resources
      .filter((r) => r.meta?.name?.kind && SEARCHABLE_KINDS.has(r.meta.name.kind as ResourceKind))
      .map((r) => {
        const kind = r.meta!.name!.kind! as ResourceKind;
        const name = r.meta!.name!.name!;
        const type = RESOURCE_KIND_TO_TYPE[kind];
        return {
          name,
          type,
          projectName,
          orgName,
          route: buildRoute(type, orgName, projectName, name),
        };
      });
  } catch {
    // Silently skip projects that fail (permissions, runtime unavailable, etc.)
    return [];
  }
}

/**
 * Prefetches resources across multiple projects, in batches of 5.
 * Calls `onProgress` with accumulated results as each batch completes,
 * enabling progressive rendering in the palette.
 */
export async function prefetchAllResources(
  queryClient: QueryClient,
  orgName: string,
  projectNames: string[],
  onProgress: (items: SearchableItem[]) => void,
): Promise<void> {
  const allItems: SearchableItem[] = [];
  const capped = projectNames.slice(0, 20);

  for (let i = 0; i < capped.length; i += BATCH_SIZE) {
    const batch = capped.slice(i, i + BATCH_SIZE);
    const batchResults = await Promise.all(
      batch.map((name) => fetchProjectResources(queryClient, orgName, name)),
    );
    allItems.push(...batchResults.flat());
    onProgress([...allItems]);
  }
}
```

- [ ] **Step 2: Wire prefetch into CommandPalette.svelte**

In `CommandPalette.svelte`, add the following after the `buildProjectItems` function:

```typescript
import { onMount } from "svelte";
import { useQueryClient } from "@tanstack/svelte-query";
import { prefetchAllResources } from "./resource-prefetch";

const queryClient = useQueryClient();

// When the project list loads, start prefetching resources
$: if ($projectListQuery.data?.projects) {
  const names = $projectListQuery.data.projects
    .filter((p) => p.name)
    .map((p) => p.name!);
  prefetchAllResources(queryClient, orgName, names, (items) => {
    resourceItems = items;
  });
}
```

This uses the imperative `queryClient.fetchQuery` approach inside the async `prefetchAllResources` function, avoiding the Svelte 4 limitation of not being able to dynamically create N reactive query subscriptions. The `onProgress` callback progressively updates `resourceItems`, which is already wired into the reactive `searchItems` chain.

- [ ] **Step 3: Test manually**

Start the dev server (`rill devtool start cloud`), navigate to an org page, press Cmd+K, type a dashboard name. Verify:
- Project results appear immediately
- Dashboard/report/alert results appear progressively as resources load
- Typing a query that matches a dashboard in another project shows that result

- [ ] **Step 4: Commit**

```bash
git add web-admin/src/features/command-palette/resource-prefetch.ts web-admin/src/features/command-palette/CommandPalette.svelte
git commit -m "feat(command-palette): add cross-project resource prefetch with batched two-hop auth"
```

---

### Task 7: Polish and Edge Cases

**Files:**
- Modify: `web-admin/src/features/command-palette/CommandPalette.svelte`

- [ ] **Step 1: Handle loading state**

When the project list is still loading (`$projectListQuery.isLoading`), show a spinner or "Loading..." in the palette instead of "Type to search...".

- [ ] **Step 2: Handle error state**

When `$projectListQuery.isError` is true (all prefetch failed), show "Unable to load search data" in the palette instead of "No results."

- [ ] **Step 3: Handle partial loading indicator**

If the project list has more items than are currently prefetched (e.g., 45 total but only 20 loaded), show a subtle note at the bottom of results: "Searching N of M projects".

- [ ] **Step 3: Clear query on close**

Ensure `query` resets to `""` when the palette closes (both via Esc and via Cmd+K toggle). The `CommandDialog`'s `on:close` or the `open` binding change should trigger this.

- [ ] **Step 5: Handle keyboard modifier display**

The footer currently uses an inline `window.navigator` check which won't work during SSR. Move the `isMac` check to `onMount` or use a reactive variable initialized in the browser:

```typescript
import { browser } from "$app/environment";

$: isMac = browser && window.navigator.userAgent.includes("Macintosh");
$: modifierLabel = isMac ? "⌘K" : "Ctrl+K";
```

- [ ] **Step 6: Test edge cases manually**

- Open palette on org home page (no project context) — should work
- Open palette inside a project — should still work
- Search with special characters — should not error
- Navigate to a result, then Cmd+K again — palette should open fresh
- Press Esc — palette closes
- Press Cmd+K while open — palette closes

- [ ] **Step 7: Commit**

```bash
git add web-admin/src/features/command-palette/CommandPalette.svelte
git commit -m "feat(command-palette): polish loading states, error handling, edge cases, SSR safety"
```

---

### Task 8: Final Verification

- [ ] **Step 1: Run all unit tests**

Run: `npx vitest run web-admin/src/features/command-palette/`
Expected: All tests pass

- [ ] **Step 2: Run the build**

Run: `npm run build -w web-admin`
Expected: Build succeeds

- [ ] **Step 3: Run lint**

Run: `npm run quality`
Expected: No new lint errors

- [ ] **Step 4: Manual E2E smoke test**

With `rill devtool start cloud` running:
1. Navigate to org home page
2. Press Cmd+K — palette opens
3. Type a project name — project results appear
4. Type a dashboard name — dashboard results appear (if resource prefetch is working)
5. Arrow down to a result, press Enter — navigates to the correct page
6. Press Cmd+K again — palette opens fresh (query cleared)
7. Press Esc — palette closes
8. Press Ctrl+K on non-Mac (or test with `navigator.userAgent` override)

- [ ] **Step 5: Commit any final fixes**

```bash
git add -A
git commit -m "feat(command-palette): final polish and fixes"
```
