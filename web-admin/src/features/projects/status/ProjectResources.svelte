<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import Button from "web-common/src/components/button/Button.svelte";

  $: resources = createRuntimeServiceListResources(
    $runtime.instanceId,
    // All resource "kinds"
    undefined,
    {
      query: {
        select: (data) => {
          // Filter out the "ProjectParser" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== "rill.runtime.v1.ProjectParser",
          );
        },
        refetchOnMount: true,
      },
    },
  );

  function refreshResources() {
    $resources.refetch();
  }

  $: isLoadingOrRefetching = $resources.isLoading || $resources.isRefetching;
</script>

<section class="flex flex-col gap-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button type="secondary" on:click={refreshResources}>
      {#if $resources.isRefetching}
        Refreshing...
      {:else}
        Refresh
      {/if}
    </Button>
  </div>
  {#if isLoadingOrRefetching}
    <DelayedSpinner isLoading={isLoadingOrRefetching} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.isSuccess}
    <ProjectResourcesTable data={$resources.data} />
  {/if}
</section>
