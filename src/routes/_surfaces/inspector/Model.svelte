<script lang="ts">
import { getContext } from "svelte";
import { flip } from "svelte/animate";
import { slide } from "svelte/transition";
import { horizontalSlide } from "$lib/transitions"
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import RowIcon from "$lib/components/icons/RowIcon.svelte";
import JSONIcon from "$lib/components/icons/JsonIcon.svelte";
import IconButton from "$lib/components/IconButton.svelte";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";

import NavEntry from "$lib/components/collapsible-table-summary/NavEntry.svelte";
import CollapsibleTableSummary from "$lib/components/collapsible-table-summary/CollapsibleTableSummary.svelte";

import { formatCardinality } from "../../../lib/util/formatters"

import type { AppStore } from '$lib/app-store';


import {format} from "d3-format";
import {dataModelerService} from "$lib/app-store";

const store = getContext('rill:app:store') as AppStore;

const formatCount = format(',');
const formatRollupFactor = format(',r');

// FIXME
let outputView = 'row';
let whichTable = {
  row: RowTable,
  json: RawJSON
}

let innerWidth;

let showSources = false;
let showOutputs = true;

function sourceDestinationCompute(key, source, destination) {
  return source.reduce((acc,v) => acc + v[key], 0) / destination[key];

}

function computeRollup(source, destination) {
  return sourceDestinationCompute('cardinality', source, destination);
}

function computeCompression(source, destination) {
  return sourceDestinationCompute('sizeInBytes', source, destination);
}


let rollup;
let compression;
let sources;

let showDestination = true;

let currentQuery;
$: if ($store?.queries && $store?.activeAsset) currentQuery = $store.queries.find(q => q.id === $store.activeAsset.id);
$: if (currentQuery?.sources) sources = $store.sources.filter(source => currentQuery.sources.includes(source.path));
$: if (currentQuery?.cardinality && sources) rollup = computeRollup(sources, {cardinality: currentQuery.cardinality });
$: if (currentQuery?.sizeInBytes && sources) compression = computeCompression(sources, { sizeInBytes: currentQuery.sizeInBytes })

</script>

<svelte:window bind:innerWidth />

  <div>
    {#if currentQuery}
      {#if sources}
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
      {#if sources}
        <div class='cost p-4 grid justify-between' style='grid-template-columns: max-content max-content;'>
          <div style="font-weight: bold;">
            {#if rollup !== 1}{formatRollupFactor(rollup)}x{:else}no{/if} rollup
          </div>
          <div style="color: #666; text-align:right;">
            {formatCardinality(sources.reduce((acc,v) => acc + v.cardinality, 0))} ⭢
            {formatCardinality(currentQuery.cardinality)} rows
          </div>
          <div>
            {#if currentQuery.sizeInBytes}
            {#if compression !== 1}{formatRollupFactor(compression)}x{:else}no{/if} compression
            {:else}<button on:click={() => {}}>generate compression</button>{/if}
          </div>
          <div style="color: #666; text-align: right;">
            {formatCardinality(sources.reduce((acc,v) => acc + v.sizeInBytes, 0))} ⭢
            {#if currentQuery.sizeInBytes}{formatCardinality(currentQuery.sizeInBytes)}{:else}...{/if}
          </div>
        </div>
      {/if}
      <div class='source-tables pt-4 pb-4'>
        {#if sources}
          <NavEntry bind:active={showSources}>
            Sources
            <div class='italic text-gray-600' slot="contextual-information">
              {formatCount(sources.reduce((acc,v) => acc + v.cardinality, 0))} rows
            </div>
          </NavEntry>
          {#if showSources}
          <div transition:slide|local={{duration: 120 }}>
            {#each sources as { path, name, cardinality, profile, head, sizeInBytes, id} (id)}
            <div class='pl-3 pr-5 pt-1 pb-1' animate:flip transition:slide|local>
              <CollapsibleTableSummary
                icon={ParquetIcon}
                collapseWidth={240 + 120 + 16}
                emphasizeTitle={true}
                {name}
                {cardinality}
                {profile}
                {head}
                {path}
                {sizeInBytes}
              />
            </div>
          {/each}
          </div>
          {/if}
        {/if}
      </div>
      
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

      {#if currentQuery?.preview && currentQuery.preview.length}
      <div class='results-container'>
        <div class="inspector-header pt-4 pb-4 pr-4 grid items-baseline sticky top-0"  style="

          grid-template-columns: auto max-content;
        ">
          <NavEntry expanded={showOutputs} on:select-body={() => { showOutputs = !showOutputs}} on:expand={() => { showOutputs=!showOutputs }} >
            Preview
          </NavEntry>
          {#if showOutputs}
          <div class="inspector-button-row grid grid-flow-col justify-start" transition:horizontalSlide>
            <IconButton title="table" selected={outputView === 'row'} on:click={() => { outputView = 'row' }}>
              <RowIcon size={14} />
            </IconButton>
            <IconButton title="JSON" selected={outputView === 'json'} on:click={() => { outputView = 'json' }}>
              <JSONIcon size={14} />
            </IconButton>
          </div>
          {/if}
        </div>


        {#if showOutputs}
        <div class="results p-4 pt-0 mt-0">
          {#if currentQuery.preview}
            {#key currentQuery.query}
              <svelte:component this={whichTable[outputView]} data={currentQuery.preview} />
            {/key}
          {/if}
        </div>
        {/if}

      </div>
      {/if}
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