<script lang="ts">
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    useRuntimeServiceGetCatalogObject,
    useRuntimeServiceMigrateSingle,
    useRuntimeServiceTriggerRefresh,
  } from "@rilldata/web-common/runtime-client";
  import { refreshSource } from "@rilldata/web-local/lib/components/navigation/sources/refreshSource";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { getContext } from "svelte";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/layout-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import { IconButton } from "../../button";
  import Import from "../../icons/Import.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import Source from "../../icons/Source.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import WorkspaceHeader from "../core/WorkspaceHeader.svelte";

  export let id;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  $: currentSource = $persistentTableStore?.entities?.find(
    (entity) => entity.id === id
  );

  const onChangeCallback = async (e) => {
    dataModelerService.dispatch("updateTableName", [id, e.target.value]);
  };

  $: titleInput = currentSource?.name;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createSource = useRuntimeServiceMigrateSingle();

  $: getSource = useRuntimeServiceGetCatalogObject(
    runtimeInstanceId,
    currentSource?.tableName
  );

  $: connector = $getSource.data?.object?.source.connector as string;

  const onRefreshClick = async (tableName: string) => {
    try {
      await refreshSource(
        connector,
        tableName,
        $runtimeStore,
        $refreshSourceMutation,
        $createSource
      );
      // invalidate the data preview (async)
      dataModelerService.dispatch("collectTableInfo", [currentSource.id]);

      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogObjectQueryKey(
        runtimeInstanceId,
        tableName
      );
      await queryClient.invalidateQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

  function formatRefreshedOn(refreshedOn: string) {
    const date = new Date(refreshedOn);
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }
</script>

<div class="grid  items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader {...{ titleInput, onChangeCallback }} showStatus={false}>
    <svelte:fragment slot="icon">
      <Source />
    </svelte:fragment>
    <svelte:fragment slot="right">
      {#if $refreshSourceMutation.isLoading}
        Refreshing...
      {:else}
        <div class="flex items-center">
          {#if $getSource.isSuccess}
            <div class="ui-copy-muted">
              Imported on {formatRefreshedOn(
                $getSource.data?.object?.refreshedOn
              )}
            </div>
          {/if}
          {#if connector === "file"}
            <Tooltip location="bottom" distance={8}>
              <div style="transformY(-1px)">
                <IconButton
                  on:click={() => onRefreshClick(currentSource.tableName)}
                >
                  <Import size="16px" />
                </IconButton>
              </div>
              <TooltipContent slot="tooltip-content">
                Import local file to refresh source
              </TooltipContent>
            </Tooltip>
          {:else}
            <Tooltip location="bottom" distance={8}>
              <IconButton
                on:click={() => onRefreshClick(currentSource.tableName)}
              >
                <RefreshIcon size="16px" />
              </IconButton>
              <TooltipContent slot="tooltip-content">
                Refresh the source data
              </TooltipContent>
            </Tooltip>
          {/if}
        </div>
      {/if}
    </svelte:fragment>
  </WorkspaceHeader>
</div>
