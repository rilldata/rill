<script lang="ts">
    import { getContext } from "svelte";

    import type { AppStore } from '$lib/app-store';
    
    import ParquetIcon from "$lib/components/icons/Parquet.svelte";
    import CollapsibleTableSummary from  "$lib/components/column-profile/CollapsibleTableSummary.svelte";
    
    let innerWidth = 0;
    const store = getContext('rill:app:store') as AppStore;
    
    $: sortedTables = $store?.tables || [];
    $: if ($store?.tables) sortedTables = sortedTables.sort((a, b) => {
      if (a.profile.length > b.profile.length) return -1;
      return 1;
    })
    
    </script>
    
<svelte:window bind:innerWidth />

    <div class='drawer-container'>
        <!-- Drawer Handler -->
        <div class='assets'>
            <div class="grid grid-cols-3">
            {#if $store && $store.tables && sortedTables}
              {#each sortedTables as { path, name, cardinality, profile, head, sizeInBytes, id, nullCounts} ( id )}
              <div class='source-list pl-3 pr-5 pt-3 pb-1' style:width="{innerWidth / 3 - 6}px" >
                <CollapsibleTableSummary 
                  icon={ParquetIcon}
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