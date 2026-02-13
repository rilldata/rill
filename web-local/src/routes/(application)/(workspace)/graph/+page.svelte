<script lang="ts">
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { page } from "$app/stores";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";
  import type { ResourceStatusFilter } from "@rilldata/web-common/features/resource-graph/shared/types";
  import { Search, X } from "lucide-svelte";

  // Parse URL parameters using new API (kind/resource instead of seed)
  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);

  // Search and filter state
  let searchQuery = "";
  let statusFilter: ResourceStatusFilter = "all";
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <div slot="header" class="header">
    <div class="header-row">
      <div class="header-left">
        <h1>Project graph</h1>
      </div>
      <div class="header-right">
        <!-- Clear filters button -->
        {#if searchQuery || statusFilter !== "all"}
          <button
            type="button"
            class="clear-filters-btn"
            on:click={() => {
              searchQuery = "";
              statusFilter = "all";
            }}
          >
            <X size={12} />
            <span>Clear Filters</span>
          </button>
        {/if}
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
          <option value="all">All status</option>
          <option value="pending">Pending</option>
          <option value="errored">Errored</option>
        </select>
      </div>
    </div>
    <p>Visualize dependencies between sources, models, dashboards, and more.</p>
  </div>

  <div slot="body" class="graph-wrapper">
    <GraphContainer {seeds} {searchQuery} {statusFilter} />
  </div>
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
    @apply h-7 w-48 pl-8 pr-2 text-xs rounded border bg-surface-background;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }

  .search-input::placeholder {
    @apply text-fg-muted;
  }

  .clear-filters-btn {
    @apply flex items-center gap-1.5 h-7 px-2 text-xs text-fg-muted;
    @apply hover:text-fg-primary cursor-pointer transition-colors;
    @apply border-none bg-transparent;
  }

  .status-filter {
    @apply h-7 px-2.5 pr-7 text-xs rounded border bg-surface-background text-fg-primary;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
    @apply cursor-pointer appearance-none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.25rem center;
    background-repeat: no-repeat;
    background-size: 1.25rem 1.25rem;
  }
</style>
