<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
    type V1ListResourcesResponse,
    createRuntimeServiceListResources,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let currentResourceName: string | undefined;
  let isReconciling = false;

  const INITIAL_REFETCH_INTERVAL = 500; // Start at 500ms
  const MAX_REFETCH_INTERVAL = 10_000; // Cap at 10s
  const BACKOFF_FACTOR = 2; // Double each time
  let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

  $: ({ instanceId } = $runtime);

  function isResourceErrored(resource: V1Resource) {
    return !!resource.meta.reconcileError;
  }

  function isResourceReconciling(resource: V1Resource) {
    return (
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING
    );
  }

  $: resources = createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data: V1ListResourcesResponse) => ({
        ...data,
        // Filter out project parser and refresh triggers
        resources: data?.resources?.filter(
          (resource: V1Resource) =>
            resource.meta.name.kind !== ResourceKind.ProjectParser &&
            resource.meta.name.kind !== ResourceKind.RefreshTrigger,
        ),
      }),
      refetchInterval: (data) => {
        if (
          $resources?.isError ||
          data?.resources?.some(isResourceErrored) ||
          !data?.resources?.some(isResourceReconciling)
        ) {
          // Reset the interval when polling stops
          currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
          return false;
        }

        // Exponential backoff with a cap
        currentRefetchInterval = Math.min(
          currentRefetchInterval * BACKOFF_FACTOR,
          MAX_REFETCH_INTERVAL,
        );

        return currentRefetchInterval;
      },
    },
  });

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  $: isReconciling = Boolean(hasReconcilingResources);

  $: isRefreshButtonDisabled = hasReconcilingResources;

  $: if ($resources.isError) {
    eventBus.emit("notification", {
      type: "error",
      message: `Error loading resources: ${$resources.error?.message}`,
    });
  }

  function refreshAllSourcesAndModels() {
    isReconciling = false;

    void $createTrigger
      .mutateAsync({
        instanceId,
        data: {
          allSourcesModels: true,
        },
      })
      .catch((error) => {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to refresh all sources and models: ${error.message}`,
        });
      });

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    );
  }

  function refreshResource(resourceName: string) {
    currentResourceName = resourceName;
    isReconciling = false;

    void queryClient
      .invalidateQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
      )
      .catch((error) => {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to refresh ${resourceName}: ${error.message}`,
        });
        currentResourceName = undefined;
      });
  }

  let previousHasReconcilingResources = false;
  $: {
    if (!previousHasReconcilingResources && hasReconcilingResources) {
      // Starting reconciliation - show loading notification
      if (currentResourceName) {
        eventBus.emit("notification", {
          type: "loading",
          message: `Refreshing ${currentResourceName}...`,
        });
      }
    } else if (previousHasReconcilingResources && !hasReconcilingResources) {
      // Check for errors when reconciliation finishes
      if (currentResourceName) {
        const resource = $resources.data?.resources?.find(
          (r) => r.meta.name.name === currentResourceName,
        );

        if (resource?.meta.reconcileError) {
          eventBus.emit("notification", {
            type: "error",
            message: `Failed to refresh ${currentResourceName}: ${resource.meta.reconcileError}`,
          });
        } else {
          eventBus.emit("notification", {
            type: "success",
            message: `Successfully refreshed ${currentResourceName}`,
          });
        }
        currentResourceName = undefined;
      }
    }
    previousHasReconcilingResources = hasReconcilingResources;
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      on:click={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      {#if isRefreshButtonDisabled}
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
      data={$resources?.data?.resources}
      triggerRefresh={refreshResource}
      {isReconciling}
    />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
