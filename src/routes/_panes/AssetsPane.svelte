<script>
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";

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
    use:drag={{ side: 'left', minSize: 200, maxSize: 500 }} />
    <div class='assets'>
        <!-- <button on:click={async () => {
          let fileHandle;

          async function getFile() {
            // open file picker
            [fileHandle] = await window.showOpenFilePicker();
            console.log(fileHandle);
            if (fileHandle.kind === 'file') {
              // run file code
            } else if (fileHandle.kind === 'directory') {
              // run directory code
            }

          }
          getFile();
        }}>add query</button> -->

        <h3 class='pl-8 pb-3 pt-3'>Sources</h3>
        {#if $store && $store.sources}
          {#each ($store?.sources || []) as { path, name, cardinality, profile, head, sizeInBytes, id} (id)}
          <div class='pl-3 pr-3 pt-1 pb-1'>
            <SourcePreview 
              emphasizeTitle={activeQuery?.profile?.map(source => source.table).includes(path)}
              {name}
              {cardinality}
              {profile}
              {head}
              {path}
              {sizeInBytes}
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
  width: var(--left-sidebar-width, 200px);
  font-size: 12px;
}
</style>