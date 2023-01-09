<script lang="ts">
  import { page } from "$app/stores";
  import {
    useEmbeddedSources,
    useSourceNames,
  } from "@rilldata/web-common/features/sources/selectors";
  import {
    useRuntimeServicePutFileAndReconcile,
    V1CatalogEntry,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  import NavigationHeader from "@rilldata/web-local/lib/components/navigation/NavigationHeader.svelte";
  import RenameAssetModal from "@rilldata/web-local/lib/components/navigation/RenameAssetModal.svelte";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { EntityType } from "../../../lib/entity";
  import { useModelNames } from "../../models/selectors";
  import AddSourceModal from "../add-source/AddSourceModal.svelte";
  import { createModelFromSource } from "../createModel";
  import EmbeddedSourceNav from "./embedded/EmbeddedSourceNav.svelte";
  import SourceMenuItems from "./SourceMenuItems.svelte";
  import SourceTooltip from "./SourceTooltip.svelte";

  $: sourceNames = useSourceNames($runtimeStore.instanceId);
  $: modelNames = useModelNames($runtimeStore.instanceId);
  const createModelMutation = useRuntimeServicePutFileAndReconcile();

  $: sourceCatalogsQuery = useEmbeddedSources($runtimeStore?.instanceId);
  let embeddedSourceCatalogs: Array<V1CatalogEntry>;
  $: embeddedSourceCatalogs = $sourceCatalogsQuery?.data ?? [];

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
  toggleText="sources"
  tooltipText="Add a new data source"
>
  Sources
</NavigationHeader>

{#if showTables}
  <div class="pb-3" transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
    {#if $sourceNames?.data}
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
    <EmbeddedSourceNav />
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
