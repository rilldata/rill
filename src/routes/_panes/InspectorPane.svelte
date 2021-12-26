<script>
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import RowIcon from "$lib/components/icons/RowIcon.svelte";
import JSONIcon from "$lib/components/icons/JsonIcon.svelte";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";

import {format} from "d3-format";

const store = getContext('rill:app:store');

const formatCardinality = format(',');
const formatRollupFactor = format(',r')

// FIXME
let outputView = 'row';
let whichTable = {
  row: RowTable,
  json: RawJSON
}

let innerWidth;

function drag(node, params = { minSize: 400, maxSize: 800, property: `--right-sidebar-width` }) {
    let minSize_ = params?.minSize || 400;
    let maxSize_ = params?.maxSize || 800;
    let property_ = params?.property || '--right-sidebar-width';
    let moving = false;
    let xSpace = minSize_;

    node.style.cursor = "move";
    node.style.userSelect = "none";

    function mousedown() {
      moving = true;
    }

    function mousemove(e) {
      if (moving) {
        const size = innerWidth - e.pageX;
        if (size > minSize_ && size < maxSize_) {
          xSpace = size;
        }

        document.body.style.setProperty(property_, `${xSpace}px`)
      }
    }

    function mouseup() {
      moving = false;
    }

    node.addEventListener("mousedown", mousedown);
    window.addEventListener("mousemove", mousemove);
    window.addEventListener("mouseup", mouseup);
    return {
      update() {
        moving = false;
      },
    };
  }

let showSources;
let showOutputs;
let showDestination;

function computeCardinality(info) {

}

function sourceDestinationCompute(key, source, destination) {
  return source.reduce((acc,v) => acc + v[key], 0) / destination[key];

}

function computeRollup(source, destination) {
  return sourceDestinationCompute('cardinality', source, destination);
}

function computeCompression(source, destination) {
  return sourceDestinationCompute('size', source, destination);
}


let rollup;
let compression;


let currentQuery = {};
$: if (currentQuery?.cardinality && currentQuery?.profile) rollup = computeRollup(currentQuery.profile, {cardinality: currentQuery.cardinality });
$: if (currentQuery?.sizeInBytes && currentQuery?.profile) compression = computeCompression(currentQuery.profile, { size: currentQuery.sizeInBytes })
$: if ($store?.queries) currentQuery = $store.queries.find(q => q.id === $store.activeQuery);
</script>

<svelte:window bind:innerWidth />

<div class='drawer-container'>
  <!-- Drawer Handler -->
  <div class='drawer-handler w-4 absolute hover:cursor-col-resize'
    style="height: calc(100vh - var(--header-height));
    transform: translateX(-.5rem);"
  use:drag />
  <div class='inspector divide-y divide-gray-200'>
    {#if currentQuery && currentQuery.profile}
      <div class="w-full flex align-items-stretch flex-col">
        <button class="p-3 pt-2 pb-2 bg-transparent border border-black m-3 rounded-md" on:click={() => {
          const query = currentQuery.query;
          const exportFilename = currentQuery.name.replace('.sql', '.parquet');
          const path = `./export/${exportFilename}`
          store.action('exportToParquet', {query, path, id: currentQuery.id });
        }}>generate {currentQuery.name.replace('.sql', '.parquet')}</button>
      </div>
    {/if}
    {#if currentQuery && currentQuery.profile}
      <div class='cost p-4 grid grid-cols-2 justify-between'>
        <div style="font-weight: bold;">
          {#if rollup !== 1}{formatRollupFactor(rollup)}x{:else}no{/if} rollup
        </div>
        <div style="color: #666; text-align:right;">
          {formatCardinality(currentQuery.profile.reduce((acc,v) => acc + v.cardinality, 0))} ⭢
          {formatCardinality(currentQuery.cardinality)} rows
        </div>
        <div>
          {#if currentQuery.sizeInBytes}
          {#if compression !== 1}{formatRollupFactor(compression)}x{:else}no{/if} compression
          {:else}...{/if}
        </div>
        <div style="color: #666; text-align: right;">
          {formatCardinality(currentQuery.profile.reduce((acc,v) => acc + v.size, 0))} ⭢
          {#if currentQuery.sizeInBytes}{formatCardinality(currentQuery.sizeInBytes)} bytes{:else}...{/if}
        </div>
      </div>
    {/if}
    <div class='source-tables p-4'>
      {#if currentQuery && currentQuery.profile}
        <CollapsibleTitle bind:active={showSources}>
          Sources
          <svelte:fragment slot="contextual-information">
            {formatCardinality(currentQuery.profile.reduce((acc,v) => acc + v.cardinality, 0))} rows
          </svelte:fragment>
        </CollapsibleTitle>
        {#if showSources}
        <div transition:slide|local={{duration: 120 }}>
          {#each currentQuery.profile as source, i (source.table)}
            <div>
              <h4>
                <span>
                  {source.table}
                </span>
                <span>
                  {formatCardinality(source.cardinality)} row{#if source.cardinality !== 1}s{/if}
                </span>
              </h4>
              <table cellpadding="0" cellspacing="0">
              {#each source.info as column}
                <tr>
                  <td>
                  <div class="font-medium">{column.name} 
                      <span class="column-type">
                        {column.type}
                      </span>
                      <span class="font-light text-gray-500">
                      {#if column.pk === 1} (primary){:else}{/if}
                    </span></div> 
                  </td>
                  <td class='column-example'>
                    {(source.head[0][column.name] !== '' ? `${source.head[0][column.name]}` : '<empty>').slice(0,50)}
                  </td>
                </tr>
              {/each}
              </table>
            </div>
          {/each}
        </div>
        {/if}
      {/if}
    </div>
    
    <div class='source-tables p-4'>
      {#if currentQuery?.destinationProfile}
          <CollapsibleTitle bind:active={showDestination}>
            Destination
            <svelte:fragment slot='contextual-information'>
              {formatCardinality(currentQuery?.cardinality || 0)} row{#if currentQuery?.cardinality !== 1}s{/if}
            </svelte:fragment>
          </CollapsibleTitle>
        {#if showDestination}
        <div transition:slide|local={{duration: 120 }}>
              <table cellpadding="0" cellspacing="0">
              {#each currentQuery.destinationProfile as column}
                <tr>
                  <td>
                  <div class="font-medium">{column.name} 
                    <span class="column-type">
                      {column.type}
                    </span>
                    <span class="font-light text-gray-500">
                      {#if column.pk === 1} (primary){:else}{/if}
                    </span></div> 
                  </td>
                  <td class='column-example'>
                    {(currentQuery.preview[0][column.name] !== '' ? `${currentQuery.preview[0][column.name]}` : '<empty>').slice(0,50)}
                  </td>
                </tr>
              {/each}
              </table>
        </div>
        {/if}
      {/if}
    </div>

    {#if currentQuery?.preview}

    <div class='results-container'>
      <div class="inspector-header p-4 grid items-baseline sticky top-0"  style="
        transform: translateY({showOutputs ? '-9px' : '0px'});
        grid-template-columns: auto max-content;
      ">
        <CollapsibleTitle bind:active={showOutputs}>Preview</CollapsibleTitle>
        {#if showOutputs}
        <div class="inspector-button-row grid grid-flow-col justify-start">
          <button class='inspector-button' class:selected={outputView === 'row'} on:click={() => { outputView = 'row' }}>
            <RowIcon size={16} />
          </button>
          <button  class='inspector-button'  class:selected={outputView === 'json'} on:click={() => { outputView = 'json' }}>
            <JSONIcon size={16}  />
          </button>
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
    <div>
    </div>
  </div>
</div>
<style lang="postcss">

.drawer-container {
  display: grid;
  grid-template-columns: max-content auto;
  align-content: stretch;
  align-items: stretch;
}

.inspector {
  width: var(--right-sidebar-width, 450px);
  font-size: 13px;

}

.source-tables {
  @apply grid grid-flow-row gap-2;
}

.source-tables h4 {
  @apply m-0 mb-2 pt-2 font-black font-semibold grid grid-flow-col justify-between;
  font-size: 13px;
}

.column-type {
  @apply font-light text-gray-500;
  font-size: 11px;
}

.column-example {
  @apply font-light text-gray-500 font-normal;
}

table {
  width: 100%;
  font-size:13px;
  text-align: left;
  /* padding-right: .25rem; */
}

table tr td {
  vertical-align: top;
}

table tr td:first-child {
  padding-left: .5rem;
}

table tr td:last-child {
  text-align: right;
  color: #666;
  font-style: italic;
}

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>