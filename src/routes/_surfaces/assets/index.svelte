<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { AppStore } from '$lib/app-store';

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import AddIcon from "$lib/components/icons/Add.svelte";
import CollapsibleTableSummary from  "$lib/components/collapsible-table-summary/CollapsibleTableSummary.svelte";
import ContextButton from "$lib/components/collapsible-table-summary/ContextButton.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

import { drag } from '$lib/drag'
import {dataModelerService} from "$lib/app-store";

const store = getContext('rill:app:store') as AppStore;

$: activeQuery = $store && $store?.models && $store?.activeAsset ? $store.models.find(q => q.id === $store.activeAsset.id) : undefined;
let showTables = true;
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

      <header class='sticky top-0'>
        <h1  class='grid grid-flow-col justify-start gap-x-3 p-3 items-center content-center'>
          <div class='grid bg-gray-400 text-white w-5 h-5 items-center justify-center rounded'>
            R
          </div>
          <div class='font-normal'>untitled project</div>
        </h1>
      </header>

      <div style:height="80px"></div>

        <!-- <div 
          class='grid grid-flow-col justify-items-center justify-start pb-6 pt-6 gap-x-5 pl-3'
        >
          <button on:click={() => { view = 'assets' }}>
            <h3>Assets</h3>
          </button>
          <button disabled>
            <h3 class="font-normal text-gray-400 cursor-not-allowed" title="coming soon!">Pipelines</h3>
          </button>
        </div> -->
        <!-- <hr /> -->

          <div class='pl-3 pb-3 pt-3'>
            <CollapsibleSectionTitle bind:active={showTables}>
              <h4>Tables</h4>
            </CollapsibleSectionTitle>
          </div>
            {#if showTables}
              <div class="pb-6" transition:slide|local={{duration:200}}>
              {#if $store && $store.tables}
                {#each ($store.tables) as { path, tableName, cardinality, profile, head, sizeInBytes, id} (id)}
                <div animate:flip>
                  <CollapsibleTableSummary 
                    icon={ParquetIcon}
                    name={tableName}
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
        
          {#if $store && $store.models}
          <div class='pl-3 pb-3 pr-5 grid justify-between' style="grid-template-columns: auto max-content;">
            <CollapsibleSectionTitle bind:active={showModels}>
                <h4> Models</h4>
              </CollapsibleSectionTitle>
              <ContextButton tooltipText="create a new model" on:click={() => {
                dataModelerService.dispatch("addModel", [{}]);
                if (!showModels) {
                  showModels = true;
                }
              }}>
                <AddIcon />
              </ContextButton>

            </div>
            {#if showModels}
              <div class='pb-6 justify-self-end'  transition:slide={{duration:200}}>
              {#each $store.models as query, i (query.id)}

                <CollapsibleTableSummary
                  on:select={() => {
                    dataModelerService.dispatch("setActiveAsset", [query.id, 'model']);
                  }}
                  on:delete={() => {
                    dataModelerService.dispatch('deleteModel', [query.id]);
                  }}
                  icon={ModelIcon}
                  name={query.name}
                  cardinality={query.cardinality}
                  profile={query?.profile || []}
                  head={query.preview}
                  sizeInByptes={query?.sizeInBytes}
                  emphasizeTitle ={query?.id === $store?.activeAsset?.id}
                />
              {/each}
              </div>
            {/if}
          {/if}

    </div>


</div>
<style lang="postcss">
.drawer-container {
  /* height: calc(100vh - var(--header-height)); */
}

.assets {
  width: var(--left-sidebar-width, 400px);
  font-size: 12px;
}
</style>