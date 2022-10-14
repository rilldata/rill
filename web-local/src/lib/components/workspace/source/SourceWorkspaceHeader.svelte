<script lang="ts">
  import { getContext } from "svelte";
  import { useRuntimeServiceTriggerRefresh } from "web-common/src/runtime-client";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import type { PersistentTableStore } from "../../../application-state-stores/table-stores";
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

  const runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSource = useRuntimeServiceTriggerRefresh();

  const onRefreshClick = (tableName: string) => {
    $refreshSource.mutate(
      {
        instanceId: runtimeInstanceId,
        name: tableName,
      },
      {
        onError: (error) => {
          console.error(error);
        },
        onSuccess: () => {
          console.log("source refreshed successfully");
        },
      }
    );
  };
</script>

<div class="grid  items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader {...{ titleInput, onChangeCallback }} showStatus={false}>
    <svelte:fragment slot="icon">
      <Source />
    </svelte:fragment>
    <svelte:fragment slot="right">
      <Tooltip location="bottom" distance={8}>
        {#if $refreshSource.isLoading}
          Refreshing...
        {:else}
          <IconButton on:click={() => onRefreshClick(currentSource.tableName)}>
            <RefreshIcon />
          </IconButton>
        {/if}
        <TooltipContent slot="tooltip-content">
          refresh the source data
        </TooltipContent>
      </Tooltip>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
