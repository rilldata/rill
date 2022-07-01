<script lang="ts">
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import { flip } from "svelte/animate";

  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import notificationStore from "$lib/components/notifications/";

  import { uploadFilesWithDialog } from "$lib/util/file-upload";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import ColumnProfileNavEntry from "$lib/components/column-profile/ColumnProfileNavEntry.svelte";

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

  async function handleQueryEvent(tableName: string) {
    // check existing models to avoid a name conflict
    const existingNames = $persistentModelStore?.entities
      .filter((model) => model.name.includes(`query_${tableName}`))
      .map((model) => model.tableName)
      .sort();
    const nextName =
      existingNames.length === 0
        ? `query_${tableName}`
        : `query_${tableName}_${existingNames.length + 1}`;

    const response = await dataModelerService.dispatch("addModel", [
      {
        name: nextName,
        query: `select * from ${tableName}`,
      },
    ]);

    // change the active asset to the new model
    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      response.id,
    ]);

    notificationStore.send({
      message: `queried ${tableName} in workspace`,
    });
  }
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
>
  <CollapsibleSectionTitle tooltipText={"sources"} bind:active={showTables}>
    <h4 class="flex flex-row items-center gap-x-2">
      <ParquetIcon size="16px" /> Sources
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
            on:query={() => {
              handleQueryEvent(tableName);
            }}
            on:delete={() => {
              dataModelerService.dispatch("dropTable", [tableName]);
            }}
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
          </CollapsibleTableSummary>
        </div>
      {/each}
    {/if}
  </div>
{/if}
