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
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { onDestroy } from "svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let maxRefetchAttempts = 60; // 30 seconds maximum
  let refetchAttempts = 0;
  let pollInterval: ReturnType<typeof setInterval> | null = null;
  let individualRefresh = false;

  $: ({ instanceId } = $runtime);

  $: resources = createRuntimeServiceListResources(
    instanceId,
    // All resource "kinds"
    undefined,
    {
      query: {
        select: (data) => {
          // Exclude the "ProjectParser" resource and "RefreshTrigger" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== ResourceKind.ProjectParser &&
              resource.meta.name.kind !== ResourceKind.RefreshTrigger,
          );
        },
        refetchOnMount: true,
        keepPreviousData: true,
        onError: () => {
          stopPolling();
        },
      },
    },
  );

  $: isAnySourceOrModelReconciling = Boolean(
    $resources?.data?.some(
      (resource) =>
        resource.meta.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
        resource.meta.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    ),
  );

  $: hasReconcileError = Boolean(
    $resources?.data?.some((resource) => !!resource.meta.reconcileError),
  );

  function startPolling() {
    stopPolling();
    refetchAttempts = 0;

    pollInterval = setInterval(() => {
      refetchAttempts++;

      if (individualRefresh && hasReconcileError) {
        // Check if any resources are still reconciling
        const stillReconciling = $resources.data.some(
          (resource) =>
            resource.meta.reconcileStatus !==
            V1ReconcileStatus.RECONCILE_STATUS_IDLE,
        );

        if (!stillReconciling) {
          stopPolling();
        }

        // Refetch resources for latest reconcile status
        void $resources.refetch();

        individualRefresh = false;
        return;
      }

      if (refetchAttempts >= maxRefetchAttempts) {
        stopPolling();
        return;
      }

      void $resources.refetch();
    }, 500);
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
    refetchAttempts = 0;
  }

  $: if (!isAnySourceOrModelReconciling) {
    stopPolling();
  }

  function refreshAllSourcesAndModels() {
    startPolling();

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

  function refreshResource() {
    individualRefresh = true;
    startPolling();
    void $resources.refetch();
  }

  // Cleanup on component destroy
  onDestroy(() => {
    stopPolling();
  });
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      on:click={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={Boolean(pollInterval)}
    >
      {#if pollInterval}
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
  {:else if $resources.data}
    <ProjectResourcesTable
      data={$resources?.data}
      triggerRefresh={refreshResource}
    />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
