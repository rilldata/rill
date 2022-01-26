<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { AppStore } from '$lib/app-store';

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import DatasetPreview from  "$lib/components/DatasetPreview.svelte";
import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";

import { horizontalSlide } from "$lib/transitions";

import { drag } from '$lib/drag'

const store = getContext('rill:app:store') as AppStore;

$: activeQuery = $store && $store?.queries ? $store.queries.find(q => q.id === $store.activeQuery) : undefined;
let showSources = true;
let showModels = true;
let showMetrics = true;

let view = 'assets';

let container;
let containerWidth = 0;

onMount(() => {
    const observer = new ResizeObserver(entries => {
        containerWidth = container.clientWidth;
    });
    observer.observe(container);
})


</script>

<div class='drawer-container flex flex-row-reverse'>
    <!-- Drawer Handler -->
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize translate-x-2 body-height'
    use:drag={{ side: 'left', minSize: 300, maxSize: 500 }} />
    <div class='assets' bind:this={container}>

        <div 
          class='grid grid-flow-col justify-items-center justify-start pb-3 pt-3 gap-x-5 pl-3'
        >
          <button on:click={() => { view = 'assets' }}>
            <h3>Assets</h3>
          </button>
          <button disabled>
            <h3 class="font-normal text-gray-400 cursor-not-allowed" title="coming soon!">Pipelines</h3>
          </button>
        </div>
        <hr />

          <div class='pl-8 pb-3 pt-3'>
            <!-- TODO: rename sources to datasets in the code -->
            <CollapsibleTitle bind:active={showSources}>
              <h4 class='font-normal'>Datasets</h4>
            </CollapsibleTitle>
          </div>
            {#if showSources}
              <div class="pb-6" transition:slide|local={{duration:200}}>
              {#if $store && $store.sources}
                {#each ($store.sources) as { path, name, cardinality, profile, head, sizeInBytes, id} (id)}
                <div class='pl-3 pr-5 pb-1' animate:flip>
                  <DatasetPreview 
                    icon={ParquetIcon}
                    emphasizeTitle={activeQuery?.sources ? activeQuery?.sources.includes(path) : false}
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
          {/if}
        
          {#if $store && $store.queries}
            <div class='pl-8 pb-3'>
              <CollapsibleTitle bind:active={showModels}>
                <h4 class='font-normal'> Models</h4>
              </CollapsibleTitle>
            </div>
            {#if showModels}
              <div class='pb-6'  transition:slide={{duration:200}}>
              {#each $store.queries as query, i (query.id)}
                <div  class=' pl-8 pr-5 pb-1 grid content-center justify-items-stretch items-center w-full' style="grid-template-columns: 1rem auto max-content; grid-gap:.45rem; " class:font-bold={query.id === $store.activeQuery}>
                  <ModelIcon size={12} color="gray" /> <div>{query.name}</div> {#if  i === 0}<div style="font-size:10px" class='text-gray-400 italic  rounded border-dashed text-clip overflow-hidden whitespace-nowrap' transition:horizontalSlide={{duration: 350}}>preview coming soon</div>{/if}
                </div>
              {/each}
              </div>
            {/if}
          {/if}

      
        <div class='pl-8 pb-3 '>
          <CollapsibleTitle bind:active={showMetrics}>
            <h4 class='font-normal'> Metrics</h4>
          </CollapsibleTitle>
        </div>
        {#if showMetrics}
        <div class="pb-6 pl-12 pt-3 pr-5 italic text-gray-500" transition:slide={{duration: 200}}>
          <div class='pl-2 text-gray-400 italic '>metrics coming soon!</div>
        </div>
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