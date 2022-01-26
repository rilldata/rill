<script lang="ts">
    import { getContext } from "svelte";
    import { slide } from "svelte/transition";
    import { flip } from "svelte/animate";
    
    import type { AppStore } from '$lib/app-store';
    
    import ParquetIcon from "$lib/components/icons/Parquet.svelte";
    import DatasetPreview from  "$lib/components/DatasetPreview.svelte";
    
    let innerWidth = 0;
    const store = getContext('rill:app:store') as AppStore;
    
    $: activeQuery = $store && $store?.queries ? $store.queries.find(q => q.id === $store.activeAsset.id ) : undefined;
    $: sortedSources = $store?.sources || [];
    $: if ($store?.sources) sortedSources = sortedSources.sort((a, b) => {
      if (a.profile.length > b.profile.length) return -1;
      return 1;
    })
    
    </script>
    
<svelte:window bind:innerWidth />

    <div class='drawer-container'>
        <!-- Drawer Handler -->
        <div class='assets'>
            <div class="grid grid-cols-3">
            {#if $store && $store.sources}
              {#each sortedSources as { path, name, cardinality, profile, head, sizeInBytes, id, nullCounts} (id)}
              <div class='source-list pl-3 pr-5 pt-3 pb-1' style:width="{innerWidth / 3 - 6}px" animate:flip transition:slide|local>
                <DatasetPreview 
                  icon={ParquetIcon}
                  emphasizeTitle={activeQuery?.profile?.map(source => source.table).includes(path)}
                  {name}
                  {cardinality}
                  {profile}
                  {head}
                  {path}
                  show
                  {sizeInBytes}
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

    /* .source-list {
        overflow-y: auto;
    } */
    
    .assets {
      font-size: 12px;
      width: calc(100vw - 1rem);
    }
    </style>