<script lang="ts">
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ResourceGraph from "../embedding/ResourceGraph.svelte";
  import type {
    ResourceStatusFilter,
    ResourceStatusFilterValue,
  } from "../shared/types";

  export let seeds: string[] | undefined;
  export let searchQuery = "";
  export let statusFilter: ResourceStatusFilter = [];
  export let showSummary = true;
  export let layout: "grid" | "sidebar" = "grid";
  export let selectedGroupId: string | null = null;
  export let onSelectedGroupChange: ((id: string | null) => void) | null = null;
  export let onKindChange: ((kind: string | null) => void) | null = null;
  export let onRefreshAll: (() => void) | null = null;
  export let activeKindLabel: string = "All types";
  export let statusFilterOptions: {
    label: string;
    value: ResourceStatusFilterValue;
  }[] = [];
  export let onStatusToggle:
    | ((value: ResourceStatusFilterValue) => void)
    | null = null;
  export let onClearFilters: (() => void) | null = null;

  $: ({ instanceId } = $runtime);

  $: resourcesQuery = createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      retry: 2,
      refetchOnMount: true,
      refetchOnWindowFocus: false,
      enabled: !!instanceId,
    },
  });

  $: resources = $resourcesQuery.data?.resources ?? [];
  $: errorMessage = $resourcesQuery.error
    ? "Failed to load project resources."
    : null;
</script>

<ResourceGraph
  {resources}
  isLoading={$resourcesQuery.isLoading}
  error={errorMessage}
  {seeds}
  {searchQuery}
  {statusFilter}
  {showSummary}
  {layout}
  {selectedGroupId}
  {onSelectedGroupChange}
  {onKindChange}
  {onRefreshAll}
  {activeKindLabel}
  {statusFilterOptions}
  {onStatusToggle}
  {onClearFilters}
/>
