<script lang="ts">
    import { getContext } from "svelte";
    import { slide } from "svelte/transition";
    import { flip } from "svelte/animate";
    
    import type { AppStore } from '$lib/app-store';
    
    import ParquetIcon from "$lib/components/icons/Parquet.svelte";
    import DatasetPreview from  "$lib/components/DatasetPreview.svelte";
    
    
    const store = getContext('rill:app:store') as AppStore;
    
    $: activeQuery = $store && $store?.queries ? $store.queries.find(q => q.id === $store.activeQuery) : undefined;
    </script>
    
    <div class='drawer-container'>
        <!-- Drawer Handler -->
        <div class='assets'>
            <div class="flex">
            {#if $store && $store.sources}
              {#each ($store.sources) as { path, name, cardinality, profile, head, sizeInBytes, id, categoricalSummaries, timestampSummaries, numericalSummaries, nullCounts} (id)}
              <div style="width: 500px;" class='source-list pl-3 pr-5 pt-3 pb-1' animate:flip transition:slide|local>
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
                />
              </div>
              {/each}
            {/if}
            </div>
    
    
        </div>
    
    
    </div>
    <style lang="postcss">
    .drawer-container {
      height: calc(100vh - var(--header-height));
      overflow-y: auto;
    }

    .source-list {
        overflow-y: auto;
        height: calc(100vh - var(--header-height));
    }
    
    .assets {
      font-size: 12px;
    }
    </style>