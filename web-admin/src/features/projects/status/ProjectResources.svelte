<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    type V1ListResourcesResponse,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient, type Query } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import {
    INITIAL_REFETCH_INTERVAL,
    MAX_REFETCH_INTERVAL,
    BACKOFF_FACTOR,
    isResourceReconciling,
  } from "../../shared/refetch-interval";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;

  let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

  $: ({ instanceId } = $runtime);

  function isResourceErrored(resource: V1Resource) {
    return !!resource.meta.reconcileError;
  }

  function calculateRefetchInterval(
    currentInterval: number,
    data: V1ListResourcesResponse,
    query: Query<V1ListResourcesResponse, HTTPError>,
  ): number | false {
    if (query.state.error) return false;
    if (!data) return INITIAL_REFETCH_INTERVAL;

    const hasErrors = data.resources.some(isResourceErrored);
    const hasReconcilingResources = data.resources.some(isResourceReconciling);

    if (hasErrors || !hasReconcilingResources) {
      currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
      return false;
    }

    currentRefetchInterval = Math.min(
      currentInterval * BACKOFF_FACTOR,
      MAX_REFETCH_INTERVAL,
    );
    return currentRefetchInterval;
  }

  $: resources = createRuntimeServiceListResources(
    instanceId,
    {
      // Ensure admins can see all resources, regardless of the security policy
      skipSecurityChecks: true,
    },
    {
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
        refetchInterval: (query) =>
          calculateRefetchInterval(
            currentRefetchInterval,
            query.state.data,
            query,
          ),
      },
    },
  );

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  $: isRefreshButtonDisabled = hasReconcilingResources;

  function refreshAllSourcesAndModels() {
    void $createTrigger
      .mutateAsync({
        instanceId,
        data: { all: true },
      })
      .then(() => {
        currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
        void queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            instanceId,
            undefined,
          ),
        });
      });
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      Refresh all sources and models
    </Button>
  </div>

  {#if $resources.isLoading}
    <DelayedSpinner isLoading={$resources.isLoading} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.data}
    <ProjectResourcesTable data={$resources?.data?.resources} />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
