<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    getRuntimeServiceListFilesQueryKey,
    useRuntimeServiceGetCatalogObject,
    useRuntimeServicePutFileAndMigrate,
    useRuntimeServiceRenameFileAndMigrate,
    useRuntimeServiceTriggerRefresh,
  } from "@rilldata/web-common/runtime-client";
  import { refreshSource } from "@rilldata/web-local/lib/components/navigation/sources/refreshSource";
  import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import { getContext } from "svelte";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
  import { IconButton } from "../../button";
  import Import from "../../icons/Import.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import Source from "../../icons/Source.svelte";
  import notifications from "../../notifications";
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

  const renameSource = useRuntimeServiceRenameFileAndMigrate();

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = currentSource.name; // resets the input
      return;
    }

    dataModelerService.dispatch("updateTableName", [id, e.target.value]);
    $renameSource.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: runtimeInstanceId,
          fromPath: `sources/${currentSource.tableName}.yaml`,
          toPath: `sources/${e.target.value}.yaml`,
        },
      },
      {
        onSuccess: () => {
          goto(`/source/${e.target.value}`, { replaceState: true });
          return queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
        },
        onError: (err) => {
          console.error(err.response.data.message);
          // reset the new table name
          dataModelerService.dispatch("updateTableName", [
            currentSource?.id,
            "",
          ]);
        },
      }
    );
  };

  $: titleInput = currentSource?.name;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createSource = useRuntimeServicePutFileAndMigrate();

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
      dataModelerService.dispatch("collectTableInfo", [currentSource?.id]);

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
