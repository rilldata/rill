<script lang="ts">
  import { getContext } from "svelte";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";

  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import Source from "$lib/components/icons/Source.svelte";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import ColumnProfileNavEntry from "$lib/components/column-profile/ColumnProfileNavEntry.svelte";
  import Cancel from "$lib/components/icons/Cancel.svelte";
  import EditIcon from "$lib/components/icons/EditIcon.svelte";
  import Explore from "$lib/components/icons/Explore.svelte";
  import Model from "$lib/components/icons/Model.svelte";
  import { Divider, MenuItem } from "$lib/components/menu";
  import RenameTableModal from "$lib/components/table/RenameTableModal.svelte";
  import {
    autoCreateMetricsDefinitionForSource,
    createModelForSource,
    deleteSourceApi,
  } from "$lib/redux-store/source/source-apis";
  import { derivedProfileEntityHasTimestampColumn } from "$lib/redux-store/source/source-selectors";
  import { uploadFilesWithDialog } from "$lib/util/file-upload";

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  let showTables = true;

  let showRenameTableModal = false;
  let renameTableID = null;
  let renameTableName = null;

  const openRenameTableModal = (tableID: string, tableName: string) => {
    showRenameTableModal = true;
    renameTableID = tableID;
    renameTableName = tableName;
  };

  const queryHandler = async (tableName: string) => {
    await createModelForSource($persistentModelStore.entities, tableName);
  };

  const quickStartMetrics = async (id: string, tableName: string) => {
    await autoCreateMetricsDefinitionForSource(
      $persistentModelStore.entities,
      $derivedTableStore.entities,
      id,
      tableName
    );
  };
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
>
  <CollapsibleSectionTitle tooltipText={"sources"} bind:active={showTables}>
    <h4 class="flex flex-row items-center gap-x-2">
      <Source size="16px" /> Sources
    </h4>
  </CollapsibleSectionTitle>

  <ContextButton
    id={"create-table-button"}
    tooltipText="import csv or parquet file as a source"
    on:click={uploadFilesWithDialog}
  >
    <AddIcon />
  </ContextButton>
</div>
{#if showTables}
  <div class="pb-6" transition:slide|local={{ duration: 200 }}>
    {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $persistentTableStore.entities as { tableName, id } (id)}
        {@const derivedTable = $derivedTableStore.entities.find(
          (t) => t["id"] === id
        )}
        <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
          <CollapsibleTableSummary
            entityType={EntityType.Table}
            name={tableName}
            cardinality={derivedTable?.cardinality ?? 0}
            sizeInBytes={derivedTable?.sizeInBytes ?? 0}
            on:delete={() => deleteSourceApi(tableName)}
          >
            <svelte:fragment slot="summary" let:containerWidth>
              <ColumnProfileNavEntry
                indentLevel={1}
                {containerWidth}
                cardinality={derivedTable?.cardinality ?? 0}
                profile={derivedTable?.profile ?? []}
                head={derivedTable?.preview ?? []}
              />
            </svelte:fragment>
            <svelte:fragment slot="menu-items" let:toggleMenu>
              <MenuItem icon on:select={() => queryHandler(tableName)}>
                <svelte:fragment slot="icon">
                  <Model />
                </svelte:fragment>
                autogenerate model
              </MenuItem>

              <MenuItem
                disabled={!derivedProfileEntityHasTimestampColumn(derivedTable)}
                icon
                on:select={() => quickStartMetrics(id, tableName)}
              >
                <svelte:fragment slot="icon"><Explore /></svelte:fragment>
                autogenerate dashboard
                <svelte:fragment slot="description">
                  {#if !derivedProfileEntityHasTimestampColumn(derivedTable)}
                    requires a timestamp column
                  {/if}
                </svelte:fragment>
              </MenuItem>

              <Divider />
              <MenuItem
                icon
                on:select={() => {
                  openRenameTableModal(id, tableName);
                }}
              >
                <svelte:fragment slot="icon">
                  <EditIcon />
                </svelte:fragment>
                rename...
              </MenuItem>
              <!-- FIXME: this should pop up an "are you sure?" modal -->
              <MenuItem icon on:select={() => deleteSourceApi(tableName)}>
                <svelte:fragment slot="icon">
                  <Cancel />
                </svelte:fragment>
                delete</MenuItem
              >
            </svelte:fragment>
          </CollapsibleTableSummary>
        </div>
      {/each}
    {/if}
  </div>
  <RenameTableModal
    openModal={showRenameTableModal}
    closeModal={() => (showRenameTableModal = false)}
    tableID={renameTableID}
    currentTableName={renameTableName}
  />
{/if}
