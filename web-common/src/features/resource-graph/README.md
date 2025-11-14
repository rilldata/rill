Resource Graph (DAG) — Developer Guide

Overview

- The Resource Graph renders the project dependency DAG (sources → models → metrics/dashboards) using SvelteFlow (@xyflow/svelte) with positions computed via Dagre.
- Graphs can be viewed as cards or expanded to fill the graphs area. You can deep-link to any graph by passing one or more seed parameters in the URL.

Key Components

- Viewer and overlay
  - `web-common/src/features/resource-graph/ResourceGraph.svelte`: Orchestrates groups (by metrics or by seeds), renders the grid of graphs, and shows the expanded overlay. It reacts to URL seed changes and opens the seeded graph expanded by default.
  - `web-common/src/features/resource-graph/ResourceGraphCanvas.svelte`: Wraps `SvelteFlow` and exposes small options like `enableExpand`, `showControls`, and `fillParent`.
- Graph building
  - `web-common/src/features/resource-graph/build-resource-graph.ts`: Contains the layout routine and grouping utilities:
    - `buildResourceGraph(resources)`: Returns nodes and edges positioned via Dagre.
    - `partitionResourcesByMetrics(resources)`: Default grouping by metrics views (one graph per metrics view).
    - `partitionResourcesBySeeds(resources, seeds)`: Generic seed-based grouping that traverses the DAG in both directions from each seed.

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

Rendering Details and Options

- `ResourceGraphCanvas.svelte` props:
  - `enableExpand` (default true): shows the expand button on the card.
  - `showControls`: adds SvelteFlow’s Controls inside the graph.
  - `showLock` (default true): shows the lock/interactive toggle in Controls; set to `false` for expanded graphs.
  - `fillParent`: makes the canvas fill its container’s height (used in expanded overlay).
- `ResourceGraph.svelte` handles:
  - Seeding logic and auto-expand on seeds.
  - Expanded overlay that fills the `.graph-wrapper` area on the graph page.
  - Reactions to seed changes to update the expanded view.

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
