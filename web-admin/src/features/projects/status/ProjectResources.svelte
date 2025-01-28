<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useResources } from "./selectors";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let individualRefresh = false;
  let currentResourceName: string | undefined;
  let hasStartedReconciling = false;

  const POLL_INTERVAL = 1_000;

  $: ({ instanceId } = $runtime);

  $: resources = useResources(instanceId, {
    refetchInterval: (data) => {
      if (!individualRefresh) {
        return false;
      }

      return POLL_INTERVAL;
    },
  });

  $: isAnySourceOrModelReconciling = Boolean(
    $resources?.data?.resources?.some(
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
    individualRefresh = false;
    currentResourceName = undefined;
    hasStartedReconciling = false;
  }

  $: if (
    $resources?.isError &&
    individualRefresh &&
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
    individualRefresh = false;
    currentResourceName = undefined;
  }

  function refreshAllSourcesAndModels() {
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

  function refreshResource(resourceName: string) {
    individualRefresh = true;
    currentResourceName = resourceName;
    hasStartedReconciling = false;
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
      disabled={isAnySourceOrModelReconciling}
    >
      {#if isAnySourceOrModelReconciling}
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
