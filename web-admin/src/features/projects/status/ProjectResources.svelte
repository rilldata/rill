<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
    type V1ListResourcesResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useResources } from "./selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { onNavigate } from "$app/navigation";
  import { onMount } from "svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let isPollingEnabled = false;
  let currentResourceName: string | undefined;
  let hasStartedReconciling = false;
  let isLoaded = false;

  const INITIAL_POLL_INTERVAL = 1_000;
  const MAX_POLL_INTERVAL = 10_000;
  const BACKOFF_THRESHOLD_MS = 30_000;
  let pollStartTime: number | null = null;

  $: ({ instanceId } = $runtime);

  function isResourceErrored(resource: V1Resource) {
    return resource.meta.reconcileError;
  }

  function isResourceReconciling(resource: V1Resource) {
    return (
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
      resource.meta.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING
    );
  }

  $: resources = useResources(instanceId, {
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
        !isPollingEnabled ||
        $resources?.isError ||
        data?.resources?.some(isResourceErrored)
      ) {
        pollStartTime = null;
        return false;
      }

      // Initialize poll start time if not set
      if (!pollStartTime) {
        pollStartTime = Date.now();
      }

      // Calculate time elapsed since polling started
      const elapsedTime = Date.now() - pollStartTime;

      // After threshold, gradually increase interval to MAX_POLL_INTERVAL
      if (elapsedTime > BACKOFF_THRESHOLD_MS) {
        return MAX_POLL_INTERVAL;
      }

      return INITIAL_POLL_INTERVAL;
    },
  });

  $: isAnySourceOrModelReconciling = Boolean(
    $resources?.data?.resources?.some(isResourceReconciling),
  );

  $: isRefreshButtonDisabled =
    isAnySourceOrModelReconciling || isPollingEnabled;

  $: if (isAnySourceOrModelReconciling && isPollingEnabled) {
    hasStartedReconciling = true;
  }

  $: if (
    !isAnySourceOrModelReconciling &&
    isPollingEnabled &&
    hasStartedReconciling &&
    !$resources.isFetching
  ) {
    const failedResource = $resources?.data?.resources?.find(
      (r) => r.meta.reconcileError,
    )?.meta.name.name;
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
      });
    }
    isPollingEnabled = false;
    currentResourceName = undefined;
    hasStartedReconciling = false;
  }

  $: if (
    $resources?.isError &&
    isPollingEnabled &&
    currentResourceName &&
    !$resources.isFetching
  ) {
    eventBus.emit("notification", {
      type: "error",
      message: `Failed to refresh ${currentResourceName} - ${$resources.error?.message}`,
      options: {
        persisted: true,
      },
    });
    isPollingEnabled = false;
    currentResourceName = undefined;
  }

  function refreshAllSourcesAndModels() {
    isPollingEnabled = true;
    hasStartedReconciling = false;
    pollStartTime = null;

    void $createTrigger.mutateAsync({
      instanceId,
      data: {
        allSourcesModels: true,
      },
    });

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    );
  }

  onNavigate(() => {
    if (isPollingEnabled) {
      eventBus.emit("notification", {
        type: "clear-all",
      });
    }
  });

  function refreshResource(resourceName: string) {
    isPollingEnabled = true;
    currentResourceName = resourceName;
    hasStartedReconciling = false;
    pollStartTime = null;

    eventBus.emit("notification", {
      type: "loading",
      message: `Refreshing ${currentResourceName}...`,
      options: {
        persisted: true,
      },
    });
  }

  // Flag to track when user revisits the page
  onMount(() => {
    isLoaded = true;
  });

  // Continue polling if user revisits the page and there are non-idle resources
  $: if (isLoaded && $resources.data && !isPollingEnabled) {
    const hasNonIdleResources = $resources.data.resources?.some(
      isResourceReconciling,
    );

    if (hasNonIdleResources) {
      isPollingEnabled = true;
      hasStartedReconciling = true;
      pollStartTime = null;
    }
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
    />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
