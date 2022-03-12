<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import { tweened } from "svelte/motion";
import { sineOut as easing } from "svelte/easing";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
import Spacer from "$lib/components/icons/Spacer.svelte";
import * as classes from "$lib/util/component-classes";

import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { formatInteger, formatBigNumberPercentage } from "$lib/util/formatters"

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

function tableDestinationCompute(key, table, destination) {
  let inputs = table.reduce((acc,v) => acc + v[key], 0)
  return  (destination[key]) / inputs;
}

function computeRollup(table, destination) {
  return tableDestinationCompute('cardinality', table, destination);
}

let rollup;
let tables;
// get source tables?
let sourceTableReferences;
let showColumns = true;
let showExportOptions = true;
let sourceTableNames = [];

// interface tweens for the  big numbers
let bigRollupNumber = tweened(0, { duration: 700, easing });
let inputRowCardinality = tweened(0, { duration: 200, easing });
let outputRowCardinality = tweened(0, { duration: 250, easing });

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

$: outputRowCardinalityValue = currentDerivedModel?.cardinality
// $: if (currentDerivedModel && outputRowCardinalityValue !== 0) {
//   outputRowCardinality.set(outputRowCardinalityValue)
// }
$: inputRowCardinalityValue = tables ? tables.reduce((acc, v) => acc + v.cardinality, 0) : undefined;
//$: if (inputRowCardinalityValue) inputRowCardinality.set(inputRowCardinalityValue)
$: if (inputRowCardinalityValue !== undefined && outputRowCardinalityValue !== undefined) {
  rollup = outputRowCardinalityValue / inputRowCardinalityValue;
}

function validRollup(number) {
  return rollup !== Infinity && rollup !== -Infinity &&
            !isNaN(rollup)
}

$: if (rollup) bigRollupNumber.set(rollup);


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

{#key currentModel?.id}
  <div bind:this={container}>
    {#if currentModel && currentModel.query.trim().length && tables}
    <div class:text-gray-300={currentDerivedModel?.error} class='cost p-4 text-right grid justify-items-end justify-end' style=' font-size: 16px;'>
      <Tooltip location="left" alignment="center" distance={8}>
      <div class="w-max text-right">
          {#if validRollup(rollup)}
                {#if isNaN(rollup)}
                  sdd
                {:else if rollup !== 1}
                            <span style="font-weight: bold;">
                                {formatBigNumberPercentage(rollup)}
                            
                            </span> of source rows
                {:else} <span style="font-weight: bold;">no change</span> in row count

                {/if}  
            {:else}
              &nbsp;
          {/if}
      </div>
      <TooltipContent slot='tooltip-content'>
        <div class="pt-1 pb-1 font-bold">
          the rollup percentage
        </div>
        <div style:width="240px" class="pb-1">
          the ratio of destination table rows to
          source table rows, as a percentage
        </div>
      </TooltipContent>
      </Tooltip>
      <div class='text-gray-500'  class:text-gray-300={currentDerivedModel?.error}>
        <!-- {formatInteger(~~$inputRowCardinality)} ⭢
        {formatInteger(~~$outputRowCardinality)} rows -->
        {formatInteger(inputRowCardinalityValue)} ⭢
        {formatInteger(~~outputRowCardinalityValue)} rows
      </div>
    </div>
  {/if}
  <hr />

    <div class="model-profile">
    {#if currentModel && currentModel.query.trim().length}       
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
            <svelte:fragment slot="context-button">
              <Spacer />
          </svelte:fragment>
            </ColumnProfile>
          {/each}
        </div>

        {/if}
    </div>

    {/if}
  </div>
  </div>
{/key}
<style lang="postcss">

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>
