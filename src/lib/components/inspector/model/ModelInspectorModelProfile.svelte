<script lang="ts">
  import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import ColumnProfileNavEntry from "$lib/components/column-profile/ColumnProfileNavEntry.svelte";
  import * as classes from "$lib/util/component-classes";
  import { formatInteger } from "$lib/util/formatters";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  const store = getContext("rill:app:store") as ApplicationStore;
  const queryHighlight = getContext("rill:app:query-highlight");

  // get source tables?
  let sourceTableReferences = [];
  let showColumns = true;

  /** Select the explicit ID to prevent unneeded reactive updates in currentModel */
  $: activeEntityID = $store?.activeEntity?.id;

  /** get current model */
  let currentModel: PersistentModelEntity;
  $: currentModel =
    activeEntityID && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  /** get current derived model*/
  let currentDerivedModel: DerivedModelEntity;
  $: currentDerivedModel =
    activeEntityID && $derivedModelStore?.entities
      ? $derivedModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  // get source table references.
  $: if (currentDerivedModel?.sources?.length) {
    sourceTableReferences = currentDerivedModel.sources;
  }

  // toggle state for inspector sections
  let showSourceTables = true;

  function focus(reference) {
    return () => {
      if (!currentDerivedModel?.error && reference) {
        queryHighlight.set(reference.tables);
      }
    };
  }
  function blur() {
    queryHighlight.set(undefined);
  }

  $: modelHasError = !!currentDerivedModel?.error;
</script>

<div class="model-profile">
  {#if currentModel && currentModel.query.trim().length}
    <div class="pt-4 pb-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="sources"
          bind:active={showSourceTables}
        >
          Sources
        </CollapsibleSectionTitle>
      </div>

      <!-- source tables -->
      {#if showSourceTables}
        <div transition:slide|local={{ duration: 200 }} class="mt-1">
          {#each sourceTableReferences as table}
            {@const persistentTableRef = $persistentTableStore.entities.find(
              (t) => table.name === t.tableName
            )}
            {@const derivedTableRef = $derivedTableStore.entities.find(
              (derivedTable) => derivedTable?.id === persistentTableRef?.id
            )}
            {@const correspondingTableCardinality =
              derivedTableRef?.cardinality}

            {@const sourceName =
              persistentTableRef?.tableName || "unknown source"}

            {@const sourceIsDefined = !!persistentTableRef?.tableName}

            <WithModelResultTooltip {modelHasError}>
              <div
                class="grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
                style:grid-template-columns="auto max-content"
                on:focus={focus(table)}
                on:mouseover={focus(table)}
                on:mouseleave={blur}
                on:blur={blur}
                class:text-gray-500={modelHasError}
                class:italic={modelHasError}
              >
                <div class="text-ellipsis overflow-hidden whitespace-nowrap">
                  {sourceName}
                </div>

                <div class="text-gray-500 italic">
                  <!-- is there a source table with this name and cardinality established? -->
                  {#if correspondingTableCardinality}
                    {`${formatInteger(correspondingTableCardinality)} rows` ||
                      ""}
                  {/if}
                </div>
              </div>

              <!-- tooltip content -->
              <svelte:fragment slot="tooltip-title"
                >{sourceName}</svelte:fragment
              >
              <svelte:fragment slot="tooltip-right">Source</svelte:fragment>

              <svelte:fragment slot="tooltip-description">
                {#if sourceIsDefined}
                  This source table is referenced in the model query.
                {:else}
                  Data source is not known. This is likely due to a source name
                  changing.
                {/if}
              </svelte:fragment>
            </WithModelResultTooltip>
          {/each}
        </div>
      {/if}
    </div>

    <hr />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="selected columns"
          bind:active={showColumns}
        >
          selected columns
        </CollapsibleSectionTitle>
      </div>

      {#if currentDerivedModel?.profile && showColumns}
        <div transition:slide|local={{ duration: 200 }}>
          <CollapsibleTableSummary
            entityType={EntityType.Model}
            showTitle={false}
            show={showColumns}
            name={currentModel.name}
            cardinality={currentDerivedModel?.cardinality ?? 0}
            active={currentModel?.id === $store?.activeEntity?.id}
          >
            <svelte:fragment slot="summary" let:containerWidth>
              <ColumnProfileNavEntry
                indentLevel={0}
                {containerWidth}
                cardinality={currentDerivedModel?.cardinality ?? 0}
                profile={currentDerivedModel?.profile ?? []}
                head={currentDerivedModel?.preview ?? []}
                entityId={activeEntityID}
              />
            </svelte:fragment>
          </CollapsibleTableSummary>
        </div>
      {/if}
    </div>
  {/if}
</div>
