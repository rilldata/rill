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

The graph page supports two URL parameters:

### Filter by kind

Navigate to `/graph?kind=<kind>` to show all graphs of a resource kind:

- `?kind=metrics` - Show all MetricsView graphs
- `?kind=models` - Show all Model graphs (includes Sources, as Source is deprecated)
- `?kind=dashboards` - Show all Dashboard/Explore graphs

### Show specific resources

Navigate to `/graph?resource=<name>` or `/graph?resource=<kind>:<name>`:

- `?resource=Orders` - Show resource named "Orders" (defaults to MetricsView)
- `?resource=model:clean_orders` - Show model named "clean_orders"
- `?resource=source:raw_data` - Show source named "raw_data"
- `?resource=orders&resource=revenue` - Multiple resources

**Kind aliases**: `metrics`, `model`, `source`, `dashboard`, `canvas`

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

## Architecture & Maintainability

### Modular Structure

- **config.ts**: Centralized configuration for all layout constants
- **position-cache.ts**: Cache management for node positions
- **graph-builder.ts**: Main graph construction logic (SvelteFlow + Dagre)
- **graph-traversal.ts**: BFS-based graph traversal utilities
- **seed-parser.ts**: URL parameter parsing for deep-linking

### Cache Management

The graph uses localStorage to persist:

- **Node positions**: Stable layouts across renders
- **Group assignments**: Which resources belong to which graphs
- **Group labels**: Display names for graph groups

#### Cache Versioning

Cache version is stored in `config.ts`:

```typescript
export const CACHE_VERSION = 2; // Increment to invalidate old caches
```

**When to bump version:**

- Changing layout spacing constants (DAGRE\_\*)
- Modifying node sizing logic
- Altering graph algorithms
- Making any changes affecting node positions

Old cache versions are automatically cleaned up on first load.

#### Cache Debugging

Enable debug mode in browser console:

```javascript
window.__DEBUG_RESOURCE_GRAPH = true;
```

Access cache manager for diagnostics:

```javascript
// Get cache health stats (includes size and quota info)
window.__RESOURCE_GRAPH_CACHE.getHealthStats();
// Returns: {
//   initialized: true,
//   dirty: false,
//   quotaExceeded: false,
//   positions: 45,
//   assignments: 30,
//   labels: 5,
//   totalEntries: 80,
//   estimatedSizeBytes: 102400
// }

// Export cache data
window.__RESOURCE_GRAPH_CACHE.export();

// Clear all cached data
window.__RESOURCE_GRAPH_CACHE.clearAll();

// Import cache data (for testing/migration)
window.__RESOURCE_GRAPH_CACHE.import(data);
```

#### Cache Size Management

The cache automatically manages its size to prevent localStorage quota errors:

- **Maximum size**: 4MB (configurable in `CACHE_CONFIG.MAX_SIZE_BYTES`)
- **Auto-pruning**: Removes 25% of oldest entries when limit reached
- **Quota protection**: Catches and handles QuotaExceededError gracefully
- **Prune throttling**: Max once per 5 seconds to prevent thrashing

When quota is exceeded:

1. Cache writes are disabled to prevent errors
2. Console warning is displayed
3. Existing cache is cleared automatically
4. Call `clearAll()` to re-enable writes

#### Troubleshooting Cache Issues

**Problem: Graphs show unexpected layouts or missing connections**

Solution: Clear the cache to force recalculation:

```javascript
window.__RESOURCE_GRAPH_CACHE.clearAll();
```

Then reload the page.

**Problem: Console shows "LocalStorage quota exceeded" warning**

Solution: The cache has exceeded browser limits. It will auto-clear, but you can manually clear:

```javascript
window.__RESOURCE_GRAPH_CACHE.clearAll();
```

**Problem: Graph positions keep resetting**

Possible causes:

- Private browsing mode (cache disabled)
- Browser security settings blocking localStorage
- Cache version was bumped (intentional invalidation)

Check quota status:

```javascript
const stats = window.__RESOURCE_GRAPH_CACHE.getHealthStats();
console.log("Quota exceeded:", stats.quotaExceeded);
console.log("Cache size:", stats.estimatedSizeBytes);
```

#### Cache Invalidation Scenarios

Cache is automatically invalidated when:

1. **Version mismatch**: `CACHE_VERSION` doesn't match stored version
2. **Manual clear**: User clears browser data or calls `clearAll()`
3. **Storage quota exceeded**: Browser quota limit reached, cache auto-clears
4. **Size limit exceeded**: Cache exceeds 4MB, oldest entries pruned automatically
5. **Security error**: localStorage access denied (private browsing), caching disabled

Cache persists across:

- Page reloads
- Tab closes
- Browser restarts (unless cleared)

**Private browsing mode**: Cache writes are disabled if localStorage throws SecurityError. Graphs will work but positions won't persist.

## Troubleshooting

**Empty graph?**
Only Sources, Models, MetricsView, Explore, and Canvas are shown. Hidden resources (`meta.hidden: true`) are filtered out.

**Layout issues?**

1. Clear cache: `window.__RESOURCE_GRAPH_CACHE.clearAll()`
2. Refresh page

**Seeds not working?**
Verify resource exists and use correct kind alias. Check console for warnings.

**Performance issues?**
Enable debug mode to see performance metrics:

```javascript
window.__DEBUG_RESOURCE_GRAPH = true;
```

**Cache not persisting?**
Check browser storage quota and privacy settings. Incognito mode may disable localStorage.

## Performance

### Optimization Techniques

1. **BFS with index-based iteration**: O(V + E) instead of O(V + E \* V) for array.shift()
2. **Caching**: Node positions cached to avoid re-layout
3. **Debounced writes**: Cache writes debounced to 300ms
4. **Set-based lookups**: O(1) lookups instead of O(n) arrays
5. **Lazy rendering**: Only visible elements rendered

### Performance Monitoring

Enable performance logging:

```javascript
window.__DEBUG_RESOURCE_GRAPH = true;
```

Operations are automatically profiled and logged to console.

### Recommended Limits

- **Max nodes per graph**: 200 (tested up to 500)
- **Max graphs per page**: 10 (use `maxGroups` prop)
- **Max cache entries**: 10,000 (automatic cleanup at 15,000)

## Accessibility

### Keyboard Navigation

- **Tab**: Navigate between nodes
- **Enter/Space**: Select node
- **Escape**: Close modals/overlays
- **Arrow keys**: Future: Navigate graph structure

### Screen Reader Support

- ARIA labels on all interactive elements
- Role annotations for graph structure
- Live regions for dynamic updates (coming soon)

### Focus Management

- Focus trapped in modals
- Focus restored on close
- Visible focus indicators

## Development

### Running Tests

```bash
npm test resource-graph
```

### Debug Mode

```javascript
window.__DEBUG_RESOURCE_GRAPH = true;
```

Enables:

- Operation logging
- Cache operation logging
- Performance profiling
- Internal state exposure

### Adding New Features

1. Update `config.ts` for new constants
2. Add types to appropriate files
3. Write tests for new functionality
4. Update this README with examples
5. Consider cache version bump if layout changes

For detailed examples, API reference, and advanced usage patterns, see the inline JSDoc comments in the component files.
