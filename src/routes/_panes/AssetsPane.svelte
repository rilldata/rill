<script>
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import SourcePreview from  "$lib/components/SourcePreview.svelte";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";

import { drag } from '$lib/drag'

import {format} from "d3-format";

const store = getContext('rill:app:store');

$: activeQuery = $store && $store?.queries ? $store.queries.find(q => q.id === $store.activeQuery) : undefined;

const formatCardinality = format(',');
const formatRollupFactor = format(',r')

// FIXME
let outputView = 'row';
let whichTable = {
    row: RowTable,
    json: RawJSON
}

let innerWidth;
</script>

<svelte:window bind:innerWidth />

<div class='drawer-container flex flex-row-reverse'>
    <!-- Drawer Handler -->
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize translate-x-2 body-height'
    use:drag={{ side: 'left', minSize: 300, maxSize: 500 }} />
    <div class='assets'>
        <h3 class='pl-8 pb-3 pt-3'>Sources</h3>
        {#if $store && $store.sources}
          {#each ($store.sources) as { path, name, cardinality, profile, head, sizeInBytes, id, categoricalSummaries, timestampSummaries, numericalSummaries} (id)}
          <div class='pl-3 pr-3 pt-1 pb-1' animate:flip transition:slide|local>
            <SourcePreview
              icon={ParquetIcon}
              emphasizeTitle={activeQuery?.profile?.map(source => source.table).includes(path)}
              {name}
              {cardinality}
              {profile}
              {head}
              {path}
              {sizeInBytes}
              {categoricalSummaries}
              {timestampSummaries}
              {numericalSummaries}
              on:updateFieldSummary={(evt) => {
                console.log('got em', evt.detail);
                store.action('updateFieldSummary', evt.detail);
              }}
            />
          </div>
          {/each}
        {/if}


    </div>


</div>
<style lang="postcss">
.drawer-container {
  height: calc(100vh - var(--header-height));
}

.assets {
  width: var(--left-sidebar-width, 300px);
  font-size: 12px;
}
</style>