<script lang="ts">
import { getContext, onMount, tick } from "svelte";
import { slide } from "svelte/transition";
import { tweened } from "svelte/motion";
import { sineOut as easing } from "svelte/easing";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
import Spacer from "$lib/components/icons/Spacer.svelte";
import * as classes from "$lib/util/component-classes";
import Export from "$lib/components/icons/Export.svelte";
import { onClickOutside } from "$lib/util/on-click-outside";
import Menu from "$lib/components/menu/Menu.svelte"
import MenuItem from "$lib/components/menu/MenuItem.svelte"

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
import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

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
$: if (currentDerivedModel?.sources) {
  sourceTableReferences = currentDerivedModel?.sources;
}

// map and filter these source tables.
$: if (sourceTableReferences?.length) {
    tables = sourceTableReferences.map(sourceTableReference => {
        const table = $persistentTableStore.entities.find(t => sourceTableReference.name === t.tableName);
        if (!table) return undefined;
        return $derivedTableStore.entities.find(derivedTable => derivedTable.id === table.id);
    }).filter(t => !!t);
  } else {
    tables = [];
  }


$: outputRowCardinalityValue = currentDerivedModel?.cardinality
$: if (outputRowCardinalityValue !== 0 && outputRowCardinalityValue !== undefined) {
  outputRowCardinality.set(outputRowCardinalityValue)
}
$: inputRowCardinalityValue = tables?.length ? tables.reduce((acc, v) => acc + v.cardinality, 0) : 0;
$: if (inputRowCardinalityValue !== undefined && outputRowCardinalityValue !== undefined) {
  rollup = outputRowCardinalityValue / inputRowCardinalityValue;
}

function validRollup(number) {
  return rollup !== Infinity && rollup !== -Infinity &&
            !isNaN(number)
}

$: if (rollup !== undefined && !isNaN(rollup)) bigRollupNumber.set(rollup);

// toggle state for inspector sections
let showSourceTables = true;


let container;
let containerWidth = 0;
let contextMenu;
let contextMenuOpen = false;
let menuX;
let menuY;
let clickOutsideListener;
$: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
}

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
    <div
      style:height="var(--header-height)"
      class:text-gray-300={currentDerivedModel?.error} 
      class='cost pl-4 pr-4 flex flex-row items-center gap-x-2'
      >



    <Tooltip location="left" alignment="middle" distance={16} suppress={contextMenuOpen}>
      <button
      bind:this={contextMenu}
      on:click={async (event) => {
          contextMenuOpen = !contextMenuOpen;
          menuX = event.clientX;
          menuY = event.clientY;
          if (!clickOutsideListener) {
              await tick();
              clickOutsideListener = onClickOutside(() => {
                  contextMenuOpen = false;
              }, contextMenu);
          }
      }}
      style:grid-column="left-control"
      class="
          hover:bg-gray-300
          hover:border-gray-300
          border-black
          transition-tranform 
          text-gray-500
          duration-100
          items-center
          justify-center
          border
          border-transparent
          rounded
          flex flex-row gap-x-2
          pl-4 pr-4
          pt-1 pb-1
        "
      >
    export
    <Export size="16px" />
</button>
    <TooltipContent slot="tooltip-content">
        export this model as a dataset
    </TooltipContent>
</Tooltip>

    <div class="grow text-right">
      <div class='text-gray-900 font-bold'  class:text-gray-300={currentDerivedModel?.error}>
        {#if inputRowCardinalityValue > 0}

          {formatInteger(~~outputRowCardinalityValue)} row{#if outputRowCardinalityValue !== 1}s{/if}
          {currentDerivedModel?.profile?.length} columns
        {:else}
          &nbsp;
        {/if}
      </div>
      <Tooltip location="left" alignment="center" distance={8}>
        <div class="italic text-gray-500" >
            {#if validRollup(rollup)}
                  {#if isNaN(rollup)}
                    ~
                  {:else if rollup === 0 }
                    no rows selected
                  {:else if rollup !== 1}
                                  {formatBigNumberPercentage($bigRollupNumber)}
                              of source table rows
                  {:else}no change in row count

                  {/if}  
              {:else if rollup === Infinity}
                &nbsp; {outputRowCardinalityValue} row{#if outputRowCardinalityValue !== 1}s{/if} selected
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
        {#if showSourceTables}
          {#if sourceTableReferences?.length && tables}
          <div transition:slide|local={{duration: 200}} class="mt-1">
            {#each sourceTableReferences as reference, index (reference.name)}
            {@const correspondingTableCardinality = tables[index]?.cardinality}
              <div
                class="grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-5 pr-5"
                style:grid-template-columns="auto max-content"
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
              <div class="text-ellipsis overflow-hidden whitespace-nowrap">
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
          {:else}
            <div class='pl-5 pr-5 p-1 italic text-gray-400'>
            none selected
            </div>
          {/if}
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



{#if contextMenuOpen}
<!-- place this above codemirror.-->
<div bind:this={contextMenu}>
    <FloatingElement relationship="mouse" target={{x: menuX, y:menuY}} location="left" alignment="start">
        <Menu on:escape={()=> { contextMenuOpen = false; }} on:item-select={() => { contextMenuOpen = false; }}>
            <MenuItem on:select={() => {
                const exportFilename = currentModel.name.replace('.sql', '.parquet');
                dataModelerService.dispatch('exportToParquet', [currentModel.id, exportFilename]);
            }}>
                Export as Parquet 
            </MenuItem>
            <MenuItem on:select={() => {
                const exportFilename = currentModel.name.replace('.sql', '.csv');
                dataModelerService.dispatch('exportToCsv', [currentModel.id, exportFilename]);
            }}>
                Export as CSV 
            </MenuItem>
        </Menu>
    </FloatingElement>
</div>
{/if}

<style lang="postcss">

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>
