<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

import { formatCompactInteger } from "$lib/util/formatters";
import * as classes from "$lib/util/component-classes";

import type { ApplicationStore } from "$lib/app-store";

import {format} from "d3-format";

import { formatInteger } from "$lib/util/formatters"

import {dataModelerService} from "$lib/app-store";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { DerivedTableStore, PersistentTableStore } from "$lib/tableStores";
import type { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";

const persistentTableStore = getContext('rill:app:persistent-table-store') as PersistentTableStore;
const derivedTableStore = getContext('rill:app:derived-table-store') as DerivedTableStore;
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;
const derivedModelStore = getContext('rill:app:derived-model-store') as DerivedModelStore;

const store = getContext('rill:app:store') as ApplicationStore;
const queryHighlight = getContext('rill:app:query-highlight');

const formatRollupFactor = format(',r');


function tableDestinationCompute(key, table, destination) {
  return table.reduce((acc,v) => acc + v[key], 0) / destination[key];
}

function computeRollup(table, destination) {
  return tableDestinationCompute('cardinality', table, destination);
}

function computeCompression(table, destination) {
  return tableDestinationCompute('sizeInBytes', table, destination);
}

let rollup;
let compression;
let tables;
// get source tables?
let sourceTableReferences;
let showColumns = true;
let sourceTableNames = [];

let currentModel: PersistentModelEntity;
$: currentModel = ($store?.activeEntity && $persistentModelStore?.entities) ?
    $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
let currentDerivedModel: DerivedModelEntity;
$: currentDerivedModel = ($store?.activeEntity && $derivedModelStore?.entities) ?
    $derivedModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
// get source table references.
$: if (currentDerivedModel?.sources?.length) sourceTableReferences = currentDerivedModel?.sources;
$: if (sourceTableReferences && $persistentTableStore?.entities && $derivedTableStore?.entities)
    tables = sourceTableReferences.map(sourceTableReference => {
        const table = $persistentTableStore.entities.find(t => sourceTableReference.name === t.tableName);
        if (!table) return undefined;
        return $derivedTableStore.entities.find(derivedTable => derivedTable.id === table.id);
    }).filter(t => !!t);
$: if (currentDerivedModel?.cardinality && tables) rollup = computeRollup(tables, {cardinality: currentDerivedModel.cardinality });
$: if (currentDerivedModel?.sizeInBytes && tables) compression = computeCompression(tables, { sizeInBytes: currentDerivedModel.sizeInBytes })

// toggle state for inspector sections
let showSourceTables = true;


let container;
let containerWidth = 0;

onMount(() => {
    const observer = new ResizeObserver(entries => {
        containerWidth = container.clientWidth;
    });
    observer.observe(container);
})

</script>


  <div bind:this={container}>
    {#if currentModel && currentModel.query.trim().length}
      {#if currentModel.query.trim().length}
        <div class="flex flex-row justify-center" style:height="var(--header-height)" >
          <button class="
            p-3 pt-1 pb-1
            m-2
            bg-white
            text-black
            border
            border-black
            transition-colors
            rounded-md" on:click={() => {
            const exportFilename = currentModel.name.replace('.sql', '.parquet');
            dataModelerService.dispatch('exportToParquet', [currentModel.id, exportFilename]);
          }}>export parquet</button>
          <button class="
            p-3 pt-1 pb-1
            m-2
            bg-white
            text-black
            border
            border-black
            transition-colors
            rounded-md" on:click={() => {
            const exportFilename = currentModel.name.replace('.sql', '.csv');
            dataModelerService.dispatch('exportToCsv', [currentModel.id, exportFilename]);
          }}>export csv</button>
        </div>
      {/if}
      {#if tables}
        <div class='cost p-4 grid justify-between' style='grid-template-columns: max-content max-content;'>
          <div style="font-weight: bold;">
            {#if rollup !== 1}{formatRollupFactor(rollup)}x{:else}no{/if} rollup
          </div>
          <div style="color: #666; text-align:right;">
            {formatCompactInteger(tables.reduce((acc, v) => acc + v.cardinality, 0))} ⭢
            {formatCompactInteger(currentDerivedModel.cardinality)} rows
          </div>
          <div>
            {#if currentDerivedModel.sizeInBytes}
            {#if compression !== 1}{formatRollupFactor(compression)}x{:else}no{/if} compression
            {:else}<button on:click={() => {}}>generate compression</button>{/if}
          </div>
          <div style="color: #666; text-align: right;">
            {formatCompactInteger(tables.reduce((acc, v) => acc + v.sizeInBytes, 0))} ⭢
            {#if currentDerivedModel.sizeInBytes}{formatCompactInteger(currentDerivedModel.sizeInBytes)}{:else}...{/if}
          </div>
        </div>
      {/if}

      <hr />
      
      <div class='pt-4 pb-4'>
        <div class=" pl-5 pr-5">
          <CollapsibleSectionTitle tooltipText="source tables" bind:active={showSourceTables}>
            Source Tables
          </CollapsibleSectionTitle>
        </div>
        {#if sourceTableReferences && tables && showSourceTables}
        <div transition:slide|local={{duration: 200}} class="mt-1">
          {#each sourceTableReferences as reference, index}
          {@const correspondingTableCardinality = tables[index]?.cardinality}
            <div
              class="flex justify-between  {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-5 pr-5"
              on:focus={() => {
                queryHighlight.set(reference.tables);
              }}
              on:mouseover={() => {
                queryHighlight.set(reference.tables);
            }}
              on:mouseleave={() => {
                queryHighlight.set(undefined)
              }}
              on:blur={() => {
                queryHighlight.set(undefined);
              }}
            >
            <div>
              {reference.name}
            </div>
            <div class="text-gray-500 italic">
              <!-- is there a source table with this name and cardinality established? -->
              {#if correspondingTableCardinality}
                {`${formatInteger(correspondingTableCardinality)} rows` || ''}
              {/if}
            </div>
          </div>
          {/each}
        </div>
        {/if}
      </div>
      
      <hr />
      
      <div class="pb-4 pt-4">
      <div class=" pl-5 pr-5">
        <CollapsibleSectionTitle tooltipText="source tables" bind:active={showColumns}>
          selected columns
        </CollapsibleSectionTitle>
      </div>

        {#if currentDerivedModel?.profile && showColumns}
        <div class='source-tables pt-4 pb-4' transition:slide|local={{duration: 200}}>
          {#each currentDerivedModel.profile as column}
            <ColumnProfile
              indentLevel={0}
              containerWidth={containerWidth}

              hideNullPercentage={false}
              hideRight={false}

              compactBreakpoint={350}

              name={column.name}
              type={column.type}
              summary={column.summary}
              totalRows={currentDerivedModel?.cardinality}
              nullCount={column.nullCount}
            >
            </ColumnProfile>
          {/each}
        </div>

        {/if}
    </div>

    {/if}
  </div>
<style lang="postcss">

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>
