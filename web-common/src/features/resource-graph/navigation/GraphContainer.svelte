<script lang="ts">
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
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

  $: instanceQuery = createRuntimeServiceGetInstance(
    instanceId,
    { sensitive: true },
    { query: { enabled: !!instanceId } },
  );
  $: olapConnectorName = $instanceQuery.data?.instance?.olapConnector;

  $: resourcesQuery = createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      retry: 2,
      refetchOnMount: true,
      refetchOnWindowFocus: false,
      enabled: !!instanceId,
    },
  });

  // Filter out non-OLAP connectors so only the OLAP connector appears in the graph.
  // If no explicit connector resource exists for the OLAP connector, inject a
  // synthetic one so it still appears as a node in the DAG.
  $: resources = (function (): V1Resource[] {
    const raw = $resourcesQuery.data?.resources ?? [];
    const filtered = raw.filter((r) => {
      if (r.meta?.name?.kind !== ResourceKind.Connector) return true;
      return r.meta?.name?.name === olapConnectorName;
    });

    // Ensure the OLAP connector is present; create a synthetic resource if needed
    if (
      olapConnectorName &&
      !filtered.some(
        (r) =>
          r.meta?.name?.kind === ResourceKind.Connector &&
          r.meta?.name?.name === olapConnectorName,
      )
    ) {
      const synthetic: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Connector,
            name: olapConnectorName,
          },
          reconcileStatus: "RECONCILE_STATUS_IDLE",
        },
      };
      return [synthetic, ...filtered];
    }

    return filtered;
  })();
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
