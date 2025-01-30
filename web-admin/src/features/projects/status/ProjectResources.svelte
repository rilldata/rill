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
  import { onMount } from "svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let currentResourceName: string | undefined;
  let isReconciling = false;
  let isLoaded = false;

  const REFETCH_INTERVAL = 1_000;

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
        // polling will occur when resources are reconciling and stop when they're done or if there's an error.
        if (
          $resources?.isError ||
          data?.resources?.some(isResourceErrored) ||
          !data?.resources?.some(isResourceReconciling)
        ) {
          return false;
        }

        return REFETCH_INTERVAL;
      },
    },
  });

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  $: isRefreshButtonDisabled = hasReconcilingResources;

  $: if (hasReconcilingResources) {
    isReconciling = true;
  }

  function refreshAllSourcesAndModels() {
    isReconciling = false;

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
    currentResourceName = resourceName;
    isReconciling = false;

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    );
  }

  // Track when user navigates away and revisits the page
  onMount(() => {
    isLoaded = true;
  });

  // Continue polling if user navigates away and revisits the page
  // and there are non-idle resources
  $: if (isLoaded && $resources.data) {
    const hasNonIdleResources = $resources.data.resources?.some(
      isResourceReconciling,
    );

    if (hasNonIdleResources) {
      isReconciling = true;
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
