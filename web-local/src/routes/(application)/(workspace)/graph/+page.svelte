<script lang="ts">
  import { onDestroy } from "svelte";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import GraphInspector from "@rilldata/web-common/features/resource-graph/inspector/GraphInspector.svelte";
  import { clearGraphNodeSelection } from "@rilldata/web-common/features/resource-graph/inspector/graph-inspector-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { page } from "$app/stores";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";
  import { Search } from "lucide-svelte";

  // Use a static path for workspace store (persists inspector toggle state)
  const GRAPH_WORKSPACE_KEY = "/graph";

  // Parse URL parameters using new API (kind/resource instead of seed)
  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);

  // Workspace layout store for inspector toggle
  $: workspace = workspaces.get(GRAPH_WORKSPACE_KEY);
  $: inspectorVisible = workspace.inspector.visible;

  // Search and filter state
  let searchQuery = "";
  let statusFilter: "all" | "pending" | "errored" = "all";

  // Clear selection when leaving the page
  onDestroy(() => {
    clearGraphNodeSelection();
  });
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

<WorkspaceContainer>
  <div slot="header" class="header">
    <div class="header-row">
      <div class="header-left">
        <h1>Project graph</h1>
      </div>
      <div class="header-right">
        <!-- Search input -->
        <div class="search-container">
          <Search size={14} class="search-icon" />
          <input
            type="text"
            placeholder="Search resources..."
            bind:value={searchQuery}
            class="search-input"
          />
        </div>
        <!-- Status filter dropdown -->
        <select bind:value={statusFilter} class="status-filter">
          <option value="all">All</option>
          <option value="pending">Pending</option>
          <option value="errored">Errored</option>
        </select>
        <Tooltip distance={8}>
          <Button
            type="secondary"
            square
            selected={$inspectorVisible}
            label="Toggle inspector visibility"
            onClick={workspace.inspector.toggle}
          >
            <HideSidebar open={$inspectorVisible} size="18px" />
          </Button>
          <TooltipContent slot="tooltip-content">
            <SlidingWords
              active={$inspectorVisible}
              direction="horizontal"
              reverse
            >
              inspector
            </SlidingWords>
          </TooltipContent>
        </Tooltip>
      </div>
    </div>
    <p>Visualize dependencies between sources, models, dashboards, and more.</p>
  </div>

  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    slot="body"
    class="graph-wrapper"
    on:click={(e) => {
      // Only clear if clicking directly on the wrapper (not bubbled from graph)
      if (e.target === e.currentTarget) {
        clearGraphNodeSelection();
      }
    }}
  >
    <GraphContainer {seeds} {searchQuery} {statusFilter} />
  </div>

  <Inspector slot="inspector" filePath={GRAPH_WORKSPACE_KEY}>
    <GraphInspector />
  </Inspector>
</WorkspaceContainer>

<style lang="postcss">
  .header {
    @apply px-4 pt-3 pb-2;
  }

  .header h1 {
    @apply text-lg font-semibold text-fg-primary;
  }

  .header-row {
    @apply flex items-center justify-between;
  }

  .header-right {
    @apply flex items-center gap-x-2;
  }

  .header p {
    @apply text-sm text-fg-secondary mt-1;
  }

  .graph-wrapper {
    @apply h-full w-full;
  }

  .search-container {
    @apply relative flex items-center;
  }

  .search-container :global(.search-icon) {
    @apply absolute left-2.5 text-fg-muted pointer-events-none;
  }

  .search-input {
    @apply h-7 w-48 pl-8 pr-3 text-xs rounded border bg-surface-background;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }

  .search-input::placeholder {
    @apply text-fg-muted;
  }

  .status-filter {
    @apply h-7 px-2 text-xs rounded border bg-surface-background text-fg-primary;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
    @apply cursor-pointer;
  }
</style>
