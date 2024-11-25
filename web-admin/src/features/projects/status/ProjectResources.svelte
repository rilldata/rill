<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import Button from "web-common/src/components/button/Button.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isReconciling = false;

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
        refetchInterval: isReconciling ? 500 : false,
      },
    },
  );

  $: isAnySourceOrModelReconciling = $resources?.data?.some(
    (resource) =>
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
  );

  function refreshAllSourcesAndModels() {
    isReconciling = true;

    void $createTrigger.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        allSourcesModelsFull: true,
      },
    });

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(
        $runtime.instanceId,
        // All resource "kinds"
        undefined,
      ),
    );
  }

  $: if (!isAnySourceOrModelReconciling) {
    isReconciling = false;
  }
</script>

<section class="flex flex-col gap-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      on:click={refreshAllSourcesAndModels}
      disabled={isReconciling}
    >
      {#if isReconciling}
        Refreshing...
      {:else}
        Refresh all sources and models
      {/if}
    </Button>
  </div>
  {#if $resources.isLoading}
    <DelayedSpinner isLoading={$resources.isLoading} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.isSuccess}
    <ProjectResourcesTable data={$resources.data} />
  {/if}
</section>
