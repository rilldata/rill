<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import CollapsibleTableSummary from "$lib/components/collapsible-table-summary/CollapsibleTableSummary.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

import { formatCardinality } from "$lib/util/formatters";
import * as classes from "$lib/util/component-classes";

import type { AppStore } from '$lib/app-store';

import {format} from "d3-format";

import { formatInteger } from "$lib/util/formatters"

import {dataModelerService} from "$lib/app-store";

const store = getContext('rill:app:store') as AppStore;
const queryHighlight = getContext('rill:app:query-highlight');

const formatRollupFactor = format(',r');

// FIXME
let outputView = 'row';
let whichTable = {
  row: RowTable,
  json: RawJSON
}

let innerWidth;

let showTables = false;
let showOutputs = true;

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

let showDestination = true;
let sourceTableNames = [];

let currentQuery;
$: if ($store?.models && $store?.activeAsset) currentQuery = $store.models.find(q => q.id === $store.activeAsset.id);
// get source table references.
$: if (currentQuery?.sources.length) sourceTableReferences = currentQuery?.sources;
$: if (sourceTableReferences) sourceTableNames = sourceTableReferences.length ? sourceTableReferences.map(r=>r.name) : [];
$: if (currentQuery?.sources) tables = $store.tables.filter(source => sourceTableNames.includes(source.name));
$: if (currentQuery?.cardinality && tables) rollup = computeRollup(tables, {cardinality: currentQuery.cardinality });
$: if (currentQuery?.sizeInBytes && tables) compression = computeCompression(tables, { sizeInBytes: currentQuery.sizeInBytes })

// toggle state for inspector sections
let showSourceTables = true;

</script>

<svelte:window bind:innerWidth />

  <div>
    {#if currentQuery && currentQuery.query.trim().length}
      {#if currentQuery.query.trim().length}
        <div class="grid justify-items-center" style:height="var(--header-height)" >
          <button class="
            p-3 pt-1 pb-1
            m-2
            bg-white
            text-black
            border
            border-black
            transition-colors
            rounded-md" on:click={() => {
            const exportFilename = currentQuery.name.replace('.sql', '.parquet');
            dataModelerService.dispatch('exportToParquet', [currentQuery.id, exportFilename]);
          }}>generate {currentQuery.name.replace('.sql', '.parquet')}</button>
        </div>
      {/if}
      {#if tables}
        <div class='cost p-4 grid justify-between' style='grid-template-columns: max-content max-content;'>
          <div style="font-weight: bold;">
            {#if rollup !== 1}{formatRollupFactor(rollup)}x{:else}no{/if} rollup
          </div>
          <div style="color: #666; text-align:right;">
            {formatCardinality(tables.reduce((acc, v) => acc + v.cardinality, 0))} ⭢
            {formatCardinality(currentQuery.cardinality)} rows
          </div>
          <div>
            {#if currentQuery.sizeInBytes}
            {#if compression !== 1}{formatRollupFactor(compression)}x{:else}no{/if} compression
            {:else}<button on:click={() => {}}>generate compression</button>{/if}
          </div>
          <div style="color: #666; text-align: right;">
            {formatCardinality(tables.reduce((acc, v) => acc + v.sizeInBytes, 0))} ⭢
            {#if currentQuery.sizeInBytes}{formatCardinality(currentQuery.sizeInBytes)}{:else}...{/if}
          </div>
        </div>
      {/if}

      <hr />
      
      <div class='pt-4 pb-4'>
        <div class=" pl-5 pr-5">
          <CollapsibleSectionTitle bind:active={showSourceTables}>
            Source Tables
          </CollapsibleSectionTitle>
        </div>
        {#if sourceTableReferences && showSourceTables}
        <div transition:slide|local={{duration: 200}} class="mt-1">
          {#each sourceTableReferences as reference}
          {@const correspondingTableCardinality = $store?.tables.find((table => table.name === reference.name)).cardinality}
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
      
      <div class='source-tables pt-4 pb-4'>
        {#if currentQuery?.profile}
            <CollapsibleTableSummary 
              collapseWidth={240 + 120 + 16}
              name="Destination"
              path=""
              bind:show={showDestination}
              cardinality={currentQuery?.cardinality}
              sizeInBytes={currentQuery?.sizeInBytes}
              profile={currentQuery.profile}
              head={currentQuery.preview}
              draggable={false}
            />
        {/if}
      </div>
    {/if}
    <div>
    </div>
  </div>
<style lang="postcss">

.source-tables {
  @apply grid grid-flow-row gap-2;
}

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>