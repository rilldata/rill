<script lang="ts">
import { getContext } from "svelte";
import { flip } from "svelte/animate";
import { slide } from "svelte/transition";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import RowIcon from "$lib/components/icons/RowIcon.svelte";
import JSONIcon from "$lib/components/icons/JsonIcon.svelte";
import IconButton from "$lib/components/IconButton.svelte";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";
import DatasetPreview from "$lib/components/DatasetPreview.svelte";

import { formatCardinality } from "../../util/formatters"

import type { AppStore } from '$lib/app-store';
import type { Query } from "../../types";

import { drag } from "$lib/drag";

import {format} from "d3-format";

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
let showOutputs;

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

let currentQuery = {} as Query;
$: if ($store?.queries) currentQuery = $store.queries.find(q => q.id === $store.activeQuery);
$: if (currentQuery?.sources) sources = $store.sources.filter(source => currentQuery.sources.includes(source.path));
$: if (currentQuery?.cardinality && sources) rollup = computeRollup(sources, {cardinality: currentQuery.cardinality });
$: if (currentQuery?.sizeInBytes && sources) compression = computeCompression(sources, { sizeInBytes: currentQuery.sizeInBytes })

</script>

<svelte:window bind:innerWidth />

<div class='drawer-container flex'>
  <!-- Drawer Handler -->

  <div class='drawer-handler w-4 absolute hover:cursor-col-resize -translate-x-2 body-height'
  use:drag={{ minSize: 400 }} />

  <div class='inspector divide-y divide-gray-200'>
    {#if currentQuery}
      {#if sources}
        <!-- <div class="w-full flex align-items-stretch flex-col">
          <button class="p-3 pt-2 pb-2 bg-transparent border border-black m-3 rounded-md" on:click={() => {
            const query = currentQuery.query;
            const exportFilename = currentQuery.name.replace('.sql', '.parquet');
            const path = `./export/${exportFilename}`
            store.action('exportToParquet', {query, path, id: currentQuery.id });
          }}>generate {currentQuery.name.replace('.sql', '.parquet')}</button>
        </div> -->
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
            {:else}...{/if}
          </div>
          <div style="color: #666; text-align: right;">
            {formatCardinality(sources.reduce((acc,v) => acc + v.sizeInBytes, 0))} ⭢
            {#if currentQuery.sizeInBytes}{formatCardinality(currentQuery.sizeInBytes)}{:else}...{/if}
          </div>
        </div>
      {/if}
      <div class='source-tables p-4'>
        {#if sources}
          <CollapsibleTitle bind:active={showSources}>
            Sources
            <div class='italic text-gray-600' slot="contextual-information">
              {formatCount(sources.reduce((acc,v) => acc + v.cardinality, 0))} rows
            </div>
          </CollapsibleTitle>
          {#if showSources}
          <div transition:slide|local={{duration: 120 }}>
            {#each sources as { path, name, cardinality, profile, head, sizeInBytes, id} (id)}
            <div class='pt-1 pb-1' animate:flip transition:slide|local>
              <DatasetPreview
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
      
      <div class='source-tables p-4'>
        {#if currentQuery?.profile}
            <DatasetPreview 
              collapseWidth={240 + 120 + 16}
              name="Destination"
              path=""
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
        <div class="inspector-header p-4 grid items-baseline sticky top-0"  style="
          transform: translateY({showOutputs ? '-6px' : '0px'});
          grid-template-columns: auto max-content;
        ">
          <CollapsibleTitle bind:active={showOutputs}>Preview</CollapsibleTitle>
          {#if showOutputs}
          <div class="inspector-button-row grid grid-flow-col justify-start">
            <IconButton title="table" selected={outputView === 'row'} on:click={() => { outputView = 'row' }}>
              <RowIcon size={16} />
            </IconButton>
            <IconButton title="JSON" selected={outputView === 'json'} on:click={() => { outputView = 'json' }}>
              <JSONIcon size={16} />
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
</div>
<style lang="postcss">

.drawer-container {
  height: calc(100vh - var(--header-height));
}

.inspector {
  width: var(--right-sidebar-width, 400px);
  font-size: 12px;
}

.source-tables {
  @apply grid grid-flow-row gap-2;
}

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>