<script lang="ts">
  import ResourcesFilterableTable from "@rilldata/web-common/features/resources/ResourcesFilterableTable.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";

  /** Pre-set status filters when navigating from the overview errors section */
  export let initialStatusFilter: string[] = [];
  /** Pre-set type filters when navigating from the overview resources section */
  export let initialTypeFilter: string[] = [];

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  let selectedStatuses: string[] = initialStatusFilter;
  let selectedTypes: string[] = initialTypeFilter;

  // React to prop changes (e.g., clicking errors section switches tab and sets filter)
  $: selectedStatuses = initialStatusFilter;
  $: selectedTypes = initialTypeFilter;

  $: resourcesQuery = createRuntimeServiceListResources(
    runtimeClient,
    {},
    { query: { refetchInterval: 5000 } },
  );

  $: resources = $resourcesQuery.data?.resources ?? [];

  $: hasReconcilingSourcesOrModels = resources.some(
    (r) =>
      (r.meta?.name?.kind === ResourceKind.Source ||
        r.meta?.name?.kind === ResourceKind.Model) &&
      (r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
        r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_RUNNING),
  );

  function refreshAllSourcesAndModels() {
    void $createTrigger.mutateAsync({ all: true }).then(() => {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    });
  }
</script>

<ResourcesFilterableTable
  {resources}
  isLoading={$resourcesQuery.isLoading}
  isError={$resourcesQuery.isError}
  errorMessage={$resourcesQuery.error?.message ?? ""}
  isRefreshDisabled={hasReconcilingSourcesOrModels}
  onRefreshAll={refreshAllSourcesAndModels}
  onRefetch={() => $resourcesQuery.refetch()}
  bind:selectedStatuses
  bind:selectedTypes
/>
