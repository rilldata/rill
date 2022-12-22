<script lang="ts">
  import { page } from "$app/stores";
  import { useRuntimeServicePutFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { useSourceNames } from "@rilldata/web-local/lib/svelte-query/sources";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { useModelNames } from "../../../svelte-query/models";
  import ColumnProfile from "../../column-profile/ColumnProfile.svelte";
  import { createModelFromSource } from "../models/createModel";
  import NavigationEntry from "../NavigationEntry.svelte";
  import NavigationHeader from "../NavigationHeader.svelte";
  import RenameAssetModal from "../RenameAssetModal.svelte";
  import AddSourceModal from "./AddSourceModal.svelte";
  import SourceMenuItems from "./SourceMenuItems.svelte";
  import SourceTooltip from "./SourceTooltip.svelte";

  $: sourceNames = useSourceNames($runtimeStore.instanceId);
  $: modelNames = useModelNames($runtimeStore.instanceId);
  const createModelMutation = useRuntimeServicePutFileAndReconcile();

  const queryClient = useQueryClient();

  let showTables = true;

  let showAddSourceModal = false;

  const openShowAddSourceModal = () => {
    showAddSourceModal = true;
  };

  const queryHandler = async (tableName: string) => {
    await createModelFromSource(
      queryClient,
      $runtimeStore.instanceId,
      $modelNames.data,
      tableName,
      $createModelMutation
    );
    // TODO: fire telemetry
  };

  let showRenameTableModal = false;
  let renameTableName = null;

  const openRenameTableModal = (tableName: string) => {
    showRenameTableModal = true;
    renameTableName = tableName;
  };
</script>

<NavigationHeader
  bind:show={showTables}
  contextButtonID={"add-table"}
  on:add={openShowAddSourceModal}
  tooltipText="Add a new data source"
  toggleText="sources"
>
  Sources
</NavigationHeader>

{#if showTables}
  <div class="pb-3" transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
    {#if $sourceNames?.data}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $sourceNames.data as sourceName (sourceName)}
        <div
          animate:flip={{ duration: 200 }}
          out:slide={{ duration: LIST_SLIDE_DURATION }}
        >
          <NavigationEntry
            href={`/source/${sourceName}`}
            open={$page.url.pathname === `/source/${sourceName}`}
            on:command-click={() => queryHandler(sourceName)}
            name={sourceName}
          >
            <svelte:fragment slot="more">
              <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
                <ColumnProfile indentLevel={1} objectName={sourceName} />
              </div>
            </svelte:fragment>

            <svelte:fragment slot="tooltip-content">
              <SourceTooltip {sourceName} />
            </svelte:fragment>

            <svelte:fragment slot="menu-items" let:toggleMenu>
              <SourceMenuItems
                {sourceName}
                {toggleMenu}
                on:rename-asset={() => {
                  openRenameTableModal(sourceName);
                }}
              />
            </svelte:fragment>
          </NavigationEntry>
        </div>
      {/each}
    {/if}
  </div>
{/if}

{#if showAddSourceModal}
  <AddSourceModal
    on:close={() => {
      showAddSourceModal = false;
    }}
  />
{/if}
{#if showRenameTableModal}
  <RenameAssetModal
    entityType={EntityType.Table}
    closeModal={() => (showRenameTableModal = false)}
    currentAssetName={renameTableName}
  />
{/if}
