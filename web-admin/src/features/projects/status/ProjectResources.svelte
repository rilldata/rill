<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshConfirmDialog from "./RefreshConfirmDialog.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let startRefetchInterval = false;
  let isRefreshConfirmDialogOpen = false;

  $: ({ instanceId } = $runtime);

  $: resources = createRuntimeServiceListResources(
    instanceId,
    // All resource "kinds"
    undefined,
    {
      query: {
        select: (data) => {
          // Exclude the "ProjectParser" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== ResourceKind.ProjectParser,
          );
        },
        refetchOnMount: true,
        refetchOnWindowFocus: true,
        refetchInterval: startRefetchInterval ? 500 : false,
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
    startRefetchInterval = true;

    void $createTrigger.mutateAsync({
      instanceId,
      data: {
        allSourcesModels: true,
      },
    });

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(
        instanceId,
        // All resource "kinds"
        undefined,
      ),
    );
  }

  $: if (!isAnySourceOrModelReconciling) {
    startRefetchInterval = false;
  }

  function triggerRefresh() {
    startRefetchInterval = true;

    void $allResources.refetch();
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      on:click={() => {
        isRefreshConfirmDialogOpen = true;
      }}
      disabled={startRefetchInterval}
    >
      {#if startRefetchInterval}
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
<<<<<<< HEAD
  {:else if $resources.isSuccess}
    <ProjectResourcesTable data={$resources?.data} />
=======
  {:else if $allResources.isSuccess}
    <ProjectResourcesTable data={$allResources?.data} {triggerRefresh} />
>>>>>>> 3f8e7d4c8 (refetch allResources on row reload click)
  {/if}
</section>

<RefreshConfirmDialog
  bind:open={isRefreshConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
