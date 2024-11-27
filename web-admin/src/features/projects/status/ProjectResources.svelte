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

  let isReconciling = false;
  let isRefreshConfirmDialogOpen = false;

  $: ({ instanceId } = $runtime);

  $: resources = createRuntimeServiceListResources(
    instanceId,
    // All resource "kinds"
    undefined,
    {
      query: {
        select: (data) => {
          // Filter out the "ProjectParser" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== ResourceKind.ProjectParser,
          );
        },
        refetchOnMount: true,
        refetchOnWindowFocus: true,
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

  function refreshSources(resourceKind: string, resourceName: string) {
    void $createTrigger.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        resources: [
          {
            kind: resourceKind,
            name: resourceName,
          },
        ],
      },
    });

    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId, {
        kind: ResourceKind.Source,
      }),
    );
  }

  $: if (!isAnySourceOrModelReconciling) {
    isReconciling = false;
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
    <ProjectResourcesTable data={$resources.data} {refreshSources} />
  {/if}
</section>

<RefreshConfirmDialog
  bind:open={isRefreshConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
