<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client/gen/runtime-service/runtime-service";
  import ResourcesTable from "./ResourcesTable.svelte";

  $: resourcesQuery = createRuntimeServiceListResources(
    $runtime.instanceId,
    {},
    {
      query: {
        refetchInterval: 5000,
      },
    },
  );

  $: resources = $resourcesQuery.data?.resources ?? [];
  $: isLoading = $resourcesQuery.isLoading;
  $: isError = $resourcesQuery.isError;
  $: error = $resourcesQuery.error;
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
  </div>

  {#if isLoading && resources.length === 0}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if isError}
    <div class="text-red-500">
      Error loading resources: {error?.message}
    </div>
  {:else}
    <ResourcesTable
      data={resources}
      onRefresh={() => $resourcesQuery.refetch()}
    />
  {/if}
</section>
