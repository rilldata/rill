# Resource Graph (DAG)

## Overview

The Resource Graph visualizes project dependencies (sources → models → metrics → dashboards) using SvelteFlow with Dagre layout. Graphs can be embedded in any component or viewed on the `/graph` page with deep-linking support.

## Quick Start

**Most common use case**: Add "View Dependencies" to any resource menu:

```svelte
<script lang="ts">
  import ResourceGraphOverlay from "./ResourceGraphOverlay.svelte";

  export let resource; // V1Resource
  export let allResources; // V1Resource[]

  let showGraph = false;
</script>

<button on:click={() => (showGraph = true)}>View Dependencies</button>

<ResourceGraphOverlay
  anchorResource={resource}
  resources={allResources}
  bind:open={showGraph}
/>
```

## Key Components

- **ResourceGraphOverlay**: Branded modal for "Quick View" from anywhere
- **ResourceGraph**: Full-featured component with URL sync and grouping
- **ResourceGraphCanvas**: Low-level visualization (wraps SvelteFlow)
- **GraphOverlay**: Generic expansion overlay (inline/fullscreen/modal modes)

## URL Deep-Linking

Navigate to `/graph?seed=<kind>:<name>` to show a specific resource's graph:

- `?seed=metrics:Orders` - Show metrics view named "Orders"
- `?seed=model:clean_orders` - Show model named "clean_orders"
- `?seed=source:raw_data` - Show source named "raw_data"
- `?seed=metrics&seed=model:orders` - Multiple graphs

**Kind aliases**: `metrics`, `model`, `source`, `dashboard`, `canvas`

**Kind expansion**: `?seed=metrics` expands to one graph per MetricsView

## Embedding in Components

**Example: Dashboard widget (no URL sync)**

```svelte
<ResourceGraph
  {resources}
  seeds={[`model:${modelName}`]}
  syncExpandedParam={false}
  showSummary={false}
  maxGroups={1}
/>
```

**Example: Sidebar mini-graph with expansion**

```svelte
<script>
  import { partitionResourcesBySeeds } from "./build-resource-graph";
  import GraphOverlay from "./GraphOverlay.svelte";

  $: groups = partitionResourcesBySeeds(resources, [`model:${name}`]);
  $: group = groups[0];

  let expanded = false;
</script>

<ResourceGraphCanvas flowId={group.id} resources={group.resources} />

<GraphOverlay
  {group}
  open={expanded}
  mode="fullscreen"
  on:close={() => (expanded = false)}
/>
```

## Common Props

**ResourceGraph**:

- `resources` - V1Resource[] (required)
- `seeds` - string[] for filtering/grouping
- `syncExpandedParam` - bool, sync expansion with URL (default: true)
- `maxGroups` - limit number of graphs shown
- `showSummary`/`showCardTitles` - toggle UI elements

**ResourceGraphCanvas**:

- `resources` - V1Resource[] (required)
- `flowId` - unique ID for this graph instance
- `showControls` - show zoom/pan controls
- `fillParent` - fill container height

**ResourceGraphOverlay**:

- `anchorResource` - resource to show graph for
- `resources` - all project resources
- `open` - visibility state
- `isLoading`/`error` - loading/error states

## Existing Integrations

Graph navigation is already integrated in:

- `features/sources/navigation/SourceMenuItems.svelte`
- `features/models/navigation/ModelMenuItems.svelte`
- `features/metrics-views/MetricsViewMenuItems.svelte`

## Troubleshooting

**Empty graph?** Only Sources, Models, MetricsView, Explore, and Canvas are shown. Hidden resources (`meta.hidden: true`) are filtered out.

**Layout issues?** Clear cache: `localStorage.removeItem('rill.resourceGraph.v1')`

**Seeds not working?** Verify resource exists and use correct kind alias. Check console for warnings.

For detailed examples, API reference, and advanced usage patterns, see the inline JSDoc comments in the component files.
