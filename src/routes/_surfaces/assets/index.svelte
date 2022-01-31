<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { AppStore } from '$lib/app-store';

import Logo from "$lib/components/Logo.svelte";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import MetricsIcon from "$lib/components/icons/List.svelte";
import DatasetPreview from  "$lib/components/DatasetPreview.svelte";
import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte"

import { horizontalSlide } from "$lib/transitions";

import { drag } from '$lib/drag'

const store = getContext('rill:app:store') as AppStore;

$: activeQuery = $store && $store?.queries && $store?.activeAsset ? $store.queries.find(q => q.id === $store.activeAsset.id) : undefined;
let showSources = true;
let showModels = true;
let showMetrics = true;
let showExplores = true;

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

      <header>
        <h1  class='grid grid-flow-col justify-start gap-x-3 p-3 items-center content-center'>
          <div class='grid bg-gray-400 text-white w-5 h-5 items-center justify-center rounded'>
            R
          </div>
          <div class='font-normal'>Data Modeler Prototype</div>
        </h1>
      </header>

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

          <div class='pl-3 pb-3 pt-3'>
            <!-- TODO: rename sources to datasets in the code -->
            <CollapsibleSectionTitle bind:active={showSources}>
              <h4>Datasets</h4>
            </CollapsibleSectionTitle>
          </div>
            {#if showSources}
              <div class="pb-6" transition:slide|local={{duration:200}}>
              {#if $store && $store.sources}
                {#each ($store.sources) as { path, name, cardinality, profile, head, sizeInBytes, id} (id)}
                <div class='pl-3 pr-5 pb-1' animate:flip>
                  <DatasetPreview 
                    icon={ParquetIcon}
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
          <div class='pl-3 pb-3 pr-8 grid' style="grid-template-columns: auto max-content;">
            <CollapsibleSectionTitle bind:active={showModels}>
                <h4> Models</h4>
              </CollapsibleSectionTitle>
              <button class='text-gray-500 italic bg-gray-100 pl-3 pr-3 rounded' style="font-size:12px;" on:click={() => {
                // FIXME: rename this action to model.
                store.action('addQuery', {});
                if (!showModels) {
                  showModels = true;
                }
              }}>new +</button>
            </div>
            {#if showModels}
              <div class='pb-6'  transition:slide={{duration:200}}>
              {#each $store.queries as query, i (query.id)}
                <button 
                  on:click={() => { store.action('setActiveAsset', { id: query.id, assetType: 'model' })}}
                  class='pl-8 pr-5 pb-1 grid justify-start justify-items-stretch items-center w-full' style="grid-template-columns: 1rem auto max-content; grid-gap:.45rem; " class:font-bold={query.id === $store?.activeAsset?.id}>
                  <ModelIcon size={12} color="gray" /> <div>{query.name}</div>
                </button>
              {/each}
              </div>
            {/if}
          {/if}

      

        <!-- Define -->
        
        <!-- <div class='pl-8 pb-3  pr-8 grid' style="grid-template-columns: auto max-content;">
          <CollapsibleTitle bind:active={showMetrics}>
            <h4 class='font-normal'> Metrics</h4>
          </CollapsibleTitle>
          <button class='text-gray-500 italic bg-gray-100 pl-3 pr-3 rounded' style="font-size:12px;" on:click={() => {
            store.action('createMetricsModel');
            if (!showMetrics) {
              showMetrics = true;
            }
          }}>new +</button>

        </div>
        {#if showMetrics}
        <div class="pl-8 pr-5 pb-6 italic" transition:slide={{duration: 200}}>
          {#each ($store?.metricsModels || []) as model (model.id)}
          <div class='grid grid-flow-col justify-items-between items-center  pb-1'>
              <button
              class:font-bold={model.id === $store?.activeAsset?.id}
              on:click={() => { store.action('setActiveAsset', { id: model.id, assetType: 'metricsDefinition' })}}
                class="grid grid-flow-col justify-start justify-items-stretch items-center gap-x-2.5"
              ><MetricsIcon size={13} /> {model.name}
            </button>
          </div>
          {/each}
        </div>
        {/if} -->



        <!-- Explore -->

        <!-- <div class='pl-8 pb-3 pr-8 grid' style="grid-template-columns: auto max-content;">
          <CollapsibleTitle bind:active={showExplores}>
            <h4 class='font-normal'> Explore</h4>
          </CollapsibleTitle>


        </div>
        {#if showExplores}
        <div class="pl-8 pr-5 italic" transition:slide={{duration: 200}}>
          {#each ($store?.exploreConfigurations || []) as explore (explore.id)}
          <div class='grid grid-flow-col justify-items-between items-center  pb-1'>
              <button
              class:font-bold={explore.id === $store?.activeAsset?.id}
              on:click={() => { store.action('setActiveAsset', { id: explore.id, assetType: 'exploreConfiguration' })}}
                class="grid grid-flow-col justify-start justify-items-stretch items-center gap-x-2.5"
              ><MetricsIcon size={13} /> {explore.name}
            </button>
          </div>
          {/each}
        </div>
        {/if} -->

    </div>


</div>
<style lang="postcss">
.drawer-container {
  /* height: calc(100vh - var(--header-height)); */
}

.assets {
  width: var(--left-sidebar-width, 300px);
  font-size: 12px;
}
</style>