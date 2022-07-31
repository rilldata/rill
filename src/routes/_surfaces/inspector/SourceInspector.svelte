<script lang="ts">
  import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ColumnProfileNavEntry from "$lib/components/column-profile/ColumnProfileNavEntry.svelte";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  const store = getContext("rill:app:store") as ApplicationStore;
  const queryHighlight = getContext("rill:app:query-highlight");

  let tables;
  // get source tables?
  let sourceTableReferences;
  let showColumns = true;

  /** Select the explicit ID to prevent unneeded reactive updates in currentTable */
  $: activeEntityID = $store?.activeEntity?.id;

  let currentTable: PersistentTableEntity;
  $: currentTable =
    activeEntityID && $persistentTableStore?.entities
      ? $persistentTableStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  let currentDerivedTable: DerivedTableEntity;
  $: currentDerivedTable =
    activeEntityID && $derivedTableStore?.entities
      ? $derivedTableStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  // get source table references.
  $: if (currentDerivedTable?.sources) {
    sourceTableReferences = currentDerivedTable?.sources;
  }

  // map and filter these source tables.
  $: if (sourceTableReferences?.length) {
    tables = sourceTableReferences
      .map((sourceTableReference) => {
        const table = $persistentTableStore.entities.find(
          (t) => sourceTableReference.name === t.tableName
        );
        if (!table) return undefined;
        return $derivedTableStore.entities.find(
          (derivedTable) => derivedTable.id === table.id
        );
      })
      .filter((t) => !!t);
  } else {
    tables = [];
  }

  // toggle state for inspector sections
  let showSourceTables = true;
</script>

<div class="table-profile">
  {#if currentTable}
    <div class="p-4">
      <div class="font-bold">
        {#if currentTable.sourceType === 0}
          parquet
        {:else if currentTable.sourceType === 1}
          <!-- CSV file. Show delimiter-->
          csv
          {currentTable.csvDelimiter || "comma"}
        {:else if currentTable.sourceType === 2}
          duckb
        {/if}
      </div>
    </div>

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="source tables"
          bind:active={showColumns}
        >
          columns
        </CollapsibleSectionTitle>
      </div>

      {#if currentDerivedTable?.profile && showColumns}
        <div transition:slide|local={{ duration: 200 }}>
          <CollapsibleTableSummary
            entityType={EntityType.Table}
            showTitle={false}
            show={showColumns}
            name={currentTable.name}
            cardinality={currentDerivedTable?.cardinality ?? 0}
            active={currentTable?.id === $store?.activeEntity?.id}
          >
            <svelte:fragment slot="summary" let:containerWidth>
              <ColumnProfileNavEntry
                indentLevel={0}
                {containerWidth}
                cardinality={currentDerivedTable?.cardinality ?? 0}
                profile={currentDerivedTable?.profile ?? []}
                head={currentDerivedTable?.preview ?? []}
              />
            </svelte:fragment>
          </CollapsibleTableSummary>
        </div>
      {/if}
    </div>
  {/if}
</div>
