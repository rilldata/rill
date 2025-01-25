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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  const POLLING_INTERVAL = 500;
  const MAX_POLLING_TIME = 30000; // 30 seconds

  let isConfirmDialogOpen = false;
  let refetchAttempts = 0;
  let pollInterval: ReturnType<typeof setInterval> | null = null;
  let individualRefresh = false;
  let currentResourceName: string | undefined;
  let hasStartedReconciling = false;

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

  $: if (isAnySourceOrModelReconciling && individualRefresh) {
    hasStartedReconciling = true;
  }

  $: if (
    !isAnySourceOrModelReconciling &&
    individualRefresh &&
    hasStartedReconciling
  ) {
    const failedResource = $resources.data?.find((r) => r.meta.reconcileError)
      ?.meta.name.name;
    if (failedResource) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to refresh ${failedResource}`,
        options: {
          persisted: true,
        },
      });
    } else if (currentResourceName) {
      eventBus.emit("notification", {
        type: "success",
        message: `Successfully refreshed ${currentResourceName}`,
        options: {
          persisted: false,
        },
      });
    }
    individualRefresh = false;
    currentResourceName = undefined;
    hasStartedReconciling = false;
    stopPolling();
  }

  function startPolling(resourceName?: string) {
    stopPolling();
    currentResourceName = resourceName;
    hasStartedReconciling = false;

    if (individualRefresh) {
      eventBus.emit("notification", {
        type: "loading",
        message: `Refreshing ${resourceName}...`,
        options: {
          persisted: true,
        },
      });
    }

    const startTime = Date.now();
    pollInterval = setInterval(() => {
      if (Date.now() - startTime > MAX_POLLING_TIME) {
        if (individualRefresh && resourceName) {
          eventBus.emit("notification", {
            type: "error",
            message: `Failed to refresh ${resourceName} (timeout)`,
          });
          individualRefresh = false;
        }
        stopPolling();
        return;
      }

      void $resources.refetch();
    }, POLLING_INTERVAL);
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

  function refreshResource(resourceName: string) {
    individualRefresh = true;
    startPolling(resourceName);
    void $resources.refetch();
  }

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
