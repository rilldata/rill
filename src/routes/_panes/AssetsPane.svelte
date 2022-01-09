<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { AppStore } from '$lib/app-store';

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import DatasetPreview from  "$lib/components/DatasetPreview.svelte";

import { drag } from '$lib/drag'

const store = getContext('rill:app:store') as AppStore;
;

$: activeQuery = $store && $store?.queries ? $store.queries.find(q => q.id === $store.activeQuery) : undefined;
</script>

<div class='drawer-container flex flex-row-reverse'>
    <!-- Drawer Handler -->
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize translate-x-2 body-height'
    use:drag={{ side: 'left', minSize: 300, maxSize: 500 }} />
    <div class='assets'>
        <h3 class='pl-8 pb-3 pt-3'>Sources</h3>
        {#if $store && $store.sources}
          {#each ($store.sources) as { path, name, cardinality, profile, head, sizeInBytes, id, categoricalSummaries, timestampSummaries, numericalSummaries, nullCounts} (id)}
          <div class='pl-3 pr-5 pt-1 pb-1' animate:flip transition:slide|local>
            <DatasetPreview 
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
              {nullCounts}
              on:updateFieldSummary={(evt) => {
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