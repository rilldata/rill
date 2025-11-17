Resource Graph (DAG) — Developer Guide

Overview

- The Resource Graph renders the project dependency DAG (sources → models → metrics/dashboards) using SvelteFlow (@xyflow/svelte) with positions computed via Dagre.
- Graphs can be viewed as cards or expanded to fill the graphs area. You can deep-link to any graph by passing one or more seed parameters in the URL.

Key Components

The resource graph feature is built with composable, reusable components:

**Core Components**:
- `ResourceGraph.svelte`: Full-featured component with URL sync, seeding, and summary graphs
- `ResourceGraphCanvas.svelte`: The visualization layer - wraps SvelteFlow with node/edge rendering
- `ResourceGraphContainer.svelte`: Data fetching wrapper for ResourceGraph

**Overlay Components** (two different overlays for different use cases):

- `ResourceGraphOverlay.svelte`: **Branded "Quick View" overlay** for viewing a resource's graph from anywhere in the app
  - Use when: Adding a "View Dependencies" action to resource menus/buttons
  - Features: Custom header with resource name, "Project Graphs" link, fixed modal size
  - Props: `anchorResource` (single resource), `resources` (all resources), `open`, `isLoading`, `error`
  - Example: Right-click menu item on a model → opens modal showing that model's graph

- `GraphOverlay.svelte`: **Generic expansion overlay** for flexible graph display modes
  - Use when: Building custom graph UIs with inline/fullscreen/modal expansion options
  - Features: Three modes (inline/fullscreen/modal), minimal UI, keyboard shortcuts
  - Props: `group` (ResourceGraphGrouping), `open`, `mode`, `showControls`, `showCloseButton`
  - Example: Expanding a graph card within ResourceGraph.svelte to fullscreen

**Data & Layout**:
- `build-resource-graph.ts`: Contains the layout routine and grouping utilities:
  - `buildResourceGraph(resources)`: Returns nodes and edges positioned via Dagre
  - `partitionResourcesByMetrics(resources)`: Default grouping by metrics views
  - `partitionResourcesBySeeds(resources, seeds)`: Seed-based grouping with DAG traversal

URL Seeds (deep links)

- Route: `/graph`
- Query parameter: `seed`
  - Repeat `seed` to render multiple graphs (one per seed).
  - Omitting a seed will fall back to metrics-based grouping.
- Seed formats
  - `kind:name` (preferred). Examples:
    - `metrics:Orders` (metrics view)
    - `model:clean_orders`
    - `source:raw_orders`
    - `dashboard:SalesOverview` (alias for Explore)
    - `canvas:ExecutiveOverview`
  - Fully qualified kinds also work: `rill.runtime.v1.MetricsView:Orders`
  - If you pass only `name` (no `:`), it defaults to a metrics view: `?seed=Orders`
- Multiple seeds example: `?seed=metrics:Orders&seed=model:clean_orders`

Kind Seeds (expand to all)

- You can pass a kind token without a name to expand into one seed per visible resource of that kind (1 graph per resource):
  - `?seed=metrics` → one graph per MetricsView
  - `?seed=sources` → one graph per Source (includes models that are defined-as-source)
  - `?seed=models` → one graph per Model
  - `?seed=dashboards` → one graph per Explore
- These tokens can be combined with explicit seeds, e.g. `?seed=metrics&seed=model:clean_orders`.
- If you actually have a metrics view named "metrics", target it explicitly with `?seed=metrics:metrics`.

Seed Aliases

- The following aliases map to runtime kinds in `ResourceGraph.svelte`:
  - `metrics`, `metric`, `metricsview` → `rill.runtime.v1.MetricsView`
  - `dashboard`, `explore` → `rill.runtime.v1.Explore`
  - `model` → `rill.runtime.v1.Model`
  - `source` → `rill.runtime.v1.Source`
  - `canvas` → `rill.runtime.v1.Canvas`

Behavior With Seeds

- When `seed` params are present, only those graphs are built (one per seed). The first seeded graph opens immediately in the expanded overlay.
- Changing seeds in-place (e.g., clicking another “View graph” link while already on `/graph`) updates the expanded overlay to the new seed.

Linking To A Seeded Graph

- Add a button or menu item and navigate to `/graph?seed=<kind>:<name>`.
- Example (Svelte):
  ```svelte
  <script lang="ts">
    import { goto } from '$app/navigation';
    function viewGraphForModel(name: string) {
      const seed = `model:${name}`;
      goto(`/graph?seed=${encodeURIComponent(seed)}`);
    }
  </script>
  ```
- You can add multiple seeds by repeating the param:
  ```ts
  const url = '/graph?seed=' + encodeURIComponent('metrics:Orders') +
              '&seed=' + encodeURIComponent('model:clean_orders');
  goto(url);
  ```

Existing “View graph” Menu Integrations

- Sources: `web-common/src/features/sources/navigation/SourceMenuItems.svelte`
- Models: `web-common/src/features/models/navigation/ModelMenuItems.svelte`
- Metrics (metrics views): `web-common/src/features/metrics-views/MetricsViewMenuItems.svelte`
- Each one builds the correct seed string and does `goto('/graph?seed=...')`.

Using Graph Components Modularly

The graph components can be used anywhere in the application. Here are practical examples:

---

### Example 1: Dashboard Widget

Display a model's dependencies in a dashboard panel without affecting the URL:

```svelte
<!-- DashboardModelCard.svelte -->
<script lang="ts">
  import ResourceGraph from '@rilldata/web-common/features/resource-graph/ResourceGraph.svelte';
  import { createRuntimeServiceListResources } from '@rilldata/web-common/runtime-client';
  import { runtime } from '@rilldata/web-common/runtime-client/runtime-store';

  export let modelName: string;

  $: instanceId = $runtime.instanceId;
  $: resourcesQuery = createRuntimeServiceListResources(instanceId);
  $: resources = $resourcesQuery.data?.resources ?? [];

  let expandedId: string | null = null;
</script>

<div class="dashboard-card">
  <h3>Dependencies: {modelName}</h3>

  <ResourceGraph
    {resources}
    seeds={[`model:${modelName}`]}
    syncExpandedParam={false}
    showSummary={false}
    showCardTitles={false}
    maxGroups={1}
    {expandedId}
    onExpandedChange={(id) => expandedId = id}
  />
</div>

<style>
  .dashboard-card {
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    padding: 16px;
    height: 400px;
  }
</style>
```

---

### Example 2a: Modal "View Dependencies" (Simple - Recommended)

Add a "View Dependencies" button using the branded ResourceGraphOverlay:

```svelte
<!-- ResourceActionsMenu.svelte -->
<script lang="ts">
  import ResourceGraphOverlay from '@rilldata/web-common/features/resource-graph/ResourceGraphOverlay.svelte';
  import type { V1Resource } from '@rilldata/web-common/runtime-client';

  export let resource: V1Resource;
  export let allResources: V1Resource[];

  let showDependencies = false;
</script>

<button on:click={() => showDependencies = true}>
  View Dependencies
</button>

<ResourceGraphOverlay
  anchorResource={resource}
  resources={allResources}
  open={showDependencies}
  bind:open={showDependencies}
/>
```

---

### Example 2b: Modal "View Dependencies" (Custom - Advanced)

For custom modal styling, use GraphOverlay with manual seed partitioning:

```svelte
<!-- CustomResourceGraphModal.svelte -->
<script lang="ts">
  import { partitionResourcesBySeeds } from '@rilldata/web-common/features/resource-graph/build-resource-graph';
  import GraphOverlay from '@rilldata/web-common/features/resource-graph/GraphOverlay.svelte';
  import type { V1Resource } from '@rilldata/web-common/runtime-client';

  export let resource: V1Resource;
  export let allResources: V1Resource[];

  let showDependencies = false;

  $: resourceKind = resource?.meta?.name?.kind?.replace('rill.runtime.v1.', '').toLowerCase();
  $: resourceName = resource?.meta?.name?.name;
  $: seed = resourceKind && resourceName ? `${resourceKind}:${resourceName}` : null;
  $: groups = seed ? partitionResourcesBySeeds(allResources, [seed]) : [];
  $: graphGroup = groups[0] ?? null;
</script>

<button on:click={() => showDependencies = true}>
  View Dependencies
</button>

{#if graphGroup}
  <GraphOverlay
    group={graphGroup}
    open={showDependencies}
    mode="modal"
    showControls={true}
    on:close={() => showDependencies = false}
  />
{/if}
```

---

### Example 3: Sidebar Mini-Graph

Show a compact dependency graph in a sidebar:

```svelte
<!-- EditorSidebar.svelte -->
<script lang="ts">
  import ResourceGraphCanvas from '@rilldata/web-common/features/resource-graph/ResourceGraphCanvas.svelte';
  import { partitionResourcesBySeeds } from '@rilldata/web-common/features/resource-graph/build-resource-graph';
  import GraphOverlay from '@rilldata/web-common/features/resource-graph/GraphOverlay.svelte';
  import type { V1Resource } from '@rilldata/web-common/runtime-client';

  export let currentResource: V1Resource;
  export let allResources: V1Resource[];

  let showExpanded = false;

  $: kind = currentResource?.meta?.name?.kind?.replace('rill.runtime.v1.', '').toLowerCase();
  $: name = currentResource?.meta?.name?.name;
  $: seed = kind && name ? `${kind}:${name}` : null;
  $: groups = seed ? partitionResourcesBySeeds(allResources, [seed]) : [];
  $: group = groups[0];
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <h4>Dependencies</h4>
    {#if group}
      <button on:click={() => showExpanded = true}>
        Expand
      </button>
    {/if}
  </div>

  {#if group}
    <div class="mini-graph">
      <ResourceGraphCanvas
        flowId={group.id}
        resources={group.resources}
        showControls={false}
        enableExpand={false}
        fillParent={false}
      />
    </div>

    <!-- Fullscreen overlay when expanded -->
    <GraphOverlay
      {group}
      open={showExpanded}
      mode="fullscreen"
      on:close={() => showExpanded = false}
    />
  {:else}
    <p class="empty">No dependencies</p>
  {/if}
</aside>

<style>
  .sidebar {
    width: 280px;
    border-left: 1px solid #e5e7eb;
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    padding: 12px;
    border-bottom: 1px solid #e5e7eb;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .mini-graph {
    flex: 1;
    min-height: 200px;
    padding: 8px;
  }
</style>
```

---

### Example 4: Resource Details Page

Show related resources on a detail page:

```svelte
<!-- ModelDetailsPage.svelte -->
<script lang="ts">
  import ResourceGraph from '@rilldata/web-common/features/resource-graph/ResourceGraph.svelte';
  import { createRuntimeServiceListResources } from '@rilldata/web-common/runtime-client';
  import { runtime } from '@rilldata/web-common/runtime-client/runtime-store';
  import { page } from '$app/stores';

  $: modelName = $page.params.model;
  $: instanceId = $runtime.instanceId;
  $: resourcesQuery = createRuntimeServiceListResources(instanceId);
  $: resources = $resourcesQuery.data?.resources ?? [];

  // Track expansion state locally (not in URL since we're already on a detail page)
  let expandedGraphId: string | null = null;
</script>

<div class="page-layout">
  <main class="content">
    <h1>{modelName}</h1>
    <!-- Other model details -->
  </main>

  <section class="dependencies-section">
    <h2>Dependency Graph</h2>

    <ResourceGraph
      {resources}
      seeds={[`model:${modelName}`]}
      syncExpandedParam={false}
      showSummary={false}
      maxGroups={1}
      expandedId={expandedGraphId}
      onExpandedChange={(id) => expandedGraphId = id}
      overlayMode="inline"
    />
  </section>
</div>

<style>
  .page-layout {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .dependencies-section {
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    padding: 16px;
  }
</style>
```

---

### Example 5: Custom Expansion Logic

Handle expansion with custom analytics or state management:

```svelte
<script lang="ts">
  import ResourceGraph from '@rilldata/web-common/features/resource-graph/ResourceGraph.svelte';
  import { writable } from 'svelte/store';

  export let resources;

  // Custom state management
  const expandedGraph = writable<string | null>(null);
  const graphInteractions = writable<Array<{ graphId: string; timestamp: number }>>([]);

  function handleGraphExpansion(id: string | null) {
    // Track analytics
    if (id) {
      console.log('Graph expanded:', id);
      graphInteractions.update(arr => [...arr, { graphId: id, timestamp: Date.now() }]);
    }

    // Update state
    expandedGraph.set(id);

    // Could also save to localStorage, send to backend, etc.
  }
</script>

<ResourceGraph
  {resources}
  expandedId={$expandedGraph}
  onExpandedChange={handleGraphExpansion}
  syncExpandedParam={false}
/>
```

---

### Example 6: Programmatic Graph Building

Build and display custom graphs programmatically:

```svelte
<script lang="ts">
  import { partitionResourcesBySeeds, buildResourceGraph } from '@rilldata/web-common/features/resource-graph/build-resource-graph';
  import ResourceGraphCanvas from '@rilldata/web-common/features/resource-graph/ResourceGraphCanvas.svelte';
  import type { V1Resource } from '@rilldata/web-common/runtime-client';

  export let resources: V1Resource[];
  export let focusResources: string[]; // e.g., ['model:orders', 'metrics:revenue']

  // Build custom graph groups
  $: graphGroups = partitionResourcesBySeeds(resources, focusResources);
</script>

<div class="graph-grid">
  {#each graphGroups as group}
    <div class="graph-card">
      <h3>{group.label}</h3>
      <p>{group.resources.length} resources</p>

      <ResourceGraphCanvas
        flowId={group.id}
        resources={group.resources}
        showControls={true}
      />
    </div>
  {/each}
</div>

<style>
  .graph-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 16px;
  }

  .graph-card {
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    padding: 16px;
  }
</style>
```

---

### Key Props for Modular Usage

**Disable URL sync:** Set `syncExpandedParam={false}` to manage state locally

**Control expansion:** Use `expandedId` + `onExpandedChange` for custom state management

**Limit graphs:** Use `maxGroups={1}` to show only the first graph

**Hide UI elements:**
- `showSummary={false}` - Hide summary counts
- `showCardTitles={false}` - Hide graph titles
- `showControls={false}` - Hide SvelteFlow controls

**Change overlay behavior:** Set `overlayMode="modal"` or `"fullscreen"` for different expansion styles

Rendering Details and Options

**ResourceGraph.svelte props:**
- `resources`: V1Resource[] - The resources to visualize
- `seeds`: string[] - Optional seeds for filtering/grouping
- `syncExpandedParam`: boolean (default: true) - Sync expansion state with URL
- `onExpandedChange`: ((id: string | null) => void) | null - Callback for expansion changes
- `expandedId`: string | null - Controlled mode: external expansion state
- `renderMode`: 'grid' | 'single' | 'list' (default: 'grid') - Layout mode
- `overlayMode`: 'inline' | 'fullscreen' | 'modal' (default: 'inline') - Expansion behavior
- `showSummary`: boolean (default: true) - Show summary counts header
- `showCardTitles`: boolean (default: true) - Show titles on graph cards
- `showControls`: boolean (default: true) - Show SvelteFlow controls in expanded view
- `enableExpansion`: boolean (default: true) - Allow graphs to be expanded
- `gridColumns`: number (default: 3) - Number of columns in grid layout
- `maxGroups`: number | null - Limit number of graphs displayed

**Slots:**
- `summary` - Replace the default summary header
- `graph-item` - Custom rendering for each graph card
- `empty-state` - Custom empty state when no graphs found

**ResourceGraphCanvas.svelte props:**
- `resources`: V1Resource[] - Resources to display in this graph
- `enableExpand` (default: true) - Show expand button on card
- `showControls`: boolean - Add SvelteFlow Controls inside the graph
- `showLock` (default: true) - Show lock/interactive toggle
- `fillParent`: boolean - Fill container's height (for expanded views)
- `titleLabel`: string | null - Optional title text
- `rootNodeIds`: string[] - Nodes to emphasize as roots

**ResourceGraphOverlay.svelte props:**
- `anchorResource`: V1Resource | undefined - The resource to show graph for
- `resources`: V1Resource[] - All resources (used to build the graph)
- `open`: boolean (default: false) - Whether overlay is visible
- `isLoading`: boolean (default: false) - Show loading state
- `error`: string | null (default: null) - Show error message

**GraphOverlay.svelte props:**
- `group`: ResourceGraphGrouping - The graph group to display
- `open`: boolean - Whether overlay is visible
- `mode`: 'inline' | 'fullscreen' | 'modal' (default: 'inline') - Display mode
- `showControls`: boolean (default: true) - Show graph controls
- `showCloseButton`: boolean (default: true) - Show close button (not for inline mode)

**Events:**
- `on:close` - Emitted when overlay should close (GraphOverlay only)

Graph Layout and Traversal

- Layout uses Dagre via `@dagrejs/dagre`.
- Traversal for seeds builds a bidirectional adjacency from `meta.refs` and collects the closure around each seed (both upstream sources and downstream dependents). See `partitionResourcesBySeeds` in `build-resource-graph.ts`.
- Group ids and labels default to the seed id and the resource’s display name.

Programmatic Usage (seeding API)

- If you need to partition graphs from code (not URL), call `partitionResourcesBySeeds` directly:
  ```ts
  import { partitionResourcesBySeeds } from './build-resource-graph';
  const groups = partitionResourcesBySeeds(resources, [
    { kind: 'rill.runtime.v1.Model', name: 'clean_orders' },
    'metrics:Orders', // strings are allowed; aliases are normalized in UI
  ]);
  ```

Troubleshooting

- Seeds not working? Verify the resource exists and the kind alias matches. You can use fully qualified kinds to be explicit.
- Expanded view doesn’t fill vertically? Ensure the containing wrapper has a concrete height. The graph page uses `.graph-wrapper` with `h-full` and the overlay body uses flex to fill available space.
- Graph doesn’t update when clicking another “View graph” link on the graph page? The component reacts to seed signature changes and resets the expanded view (see `ResourceGraph.svelte`).

Useful References

- Page route reading `seed` params: `web-local/src/routes/(application)/(workspace)/graph/+page.svelte`
- Container that fetches resources: `web-common/src/features/resource-graph/ResourceGraphContainer.svelte`
- Graph renderer and overlay: `web-common/src/features/resource-graph/ResourceGraph.svelte`
- Graph canvas and SvelteFlow setup: `web-common/src/features/resource-graph/ResourceGraphCanvas.svelte`
- Graph data + layout builders: `web-common/src/features/resource-graph/build-resource-graph.ts`
