<script lang="ts">
  import { getContext } from "svelte";
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    useRuntimeServiceGetCatalogObject,
    useRuntimeServiceTriggerRefresh,
  } from "web-common/src/runtime-client";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/layout-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import { IconButton } from "../../button";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import Source from "../../icons/Source.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import WorkspaceHeader from "../WorkspaceHeader.svelte";

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
  const refreshSource = useRuntimeServiceTriggerRefresh();

  $: getSource = useRuntimeServiceGetCatalogObject(
    runtimeInstanceId,
    currentSource?.tableName
  );

  const onRefreshClick = (tableName: string) => {
    overlay.set({ title: `Importing ${tableName}` });
    $refreshSource.mutate(
      {
        instanceId: runtimeInstanceId,
        name: tableName,
      },
      {
        onError: (error) => {
          console.error(error);
          overlay.set(null);
        },
        onSuccess: async () => {
          // invalidate the data preview (async)
          dataModelerService.dispatch("collectTableInfo", [currentSource.id]);

          // invalidate the "refreshed_on" time
          const queryKey = getRuntimeServiceGetCatalogObjectQueryKey(
            runtimeInstanceId,
            tableName
          );
          queryClient.invalidateQueries(queryKey);
          overlay.set(null);
        },
      }
    );
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
      {#if $refreshSource.isLoading}
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
        </div>
      {/if}
    </svelte:fragment>
  </WorkspaceHeader>
</div>
