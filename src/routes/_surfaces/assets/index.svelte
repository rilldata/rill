<script lang="ts">
import { getContext, onMount } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { ApplicationStore } from "$lib/app-store";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import AddIcon from "$lib/components/icons/Add.svelte";
import CollapsibleTableSummary from  "$lib/components/collapsible-table-summary/CollapsibleTableSummary.svelte";
import ContextButton from "$lib/components/collapsible-table-summary/ContextButton.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

import { drag } from '$lib/drag'
import {dataModelerService} from "$lib/app-store";
import type { DerivedTableStore, PersistentTableStore } from "$lib/tableStores";
import type { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

const store = getContext('rill:app:store') as ApplicationStore;
const persistentTableStore = getContext('rill:app:persistent-table-store') as PersistentTableStore;
const derivedTableStore = getContext('rill:app:derived-table-store') as DerivedTableStore;
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;
const derivedModelStore = getContext('rill:app:derived-model-store') as DerivedModelStore;

let activeModel: PersistentModelEntity;
$: activeModel = $store && $persistentModelStore &&
  $store?.activeEntity && $persistentModelStore?.entities ?
    $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
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
              {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
                <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
                {#each ($persistentTableStore.entities) as { path, tableName, id} (id)}
                  {@const derivedTable = $derivedTableStore.entities.find(t => t["id"] === id)}
                  <div animate:flip>
                    <CollapsibleTableSummary
                      icon={ParquetIcon}
                      name={tableName}
                      cardinality={derivedTable?.cardinality ?? 0}
                      profile={derivedTable?.profile ?? []}
                      head={derivedTable?.preview ?? []}
                      {path}
                      sizeInBytes={derivedTable?.sizeInBytes ?? 0}
                    />
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        
          {#if $persistentModelStore && $persistentModelStore.entities}
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
              <!-- TODO: fix the object property access back to m.id from m["id"] once svelte fixes it -->
              {#each $persistentModelStore.entities as query, i (query.id)}
                {@const derivedModel = $derivedModelStore.entities.find(m => m["id"] === query["id"])}
                <CollapsibleTableSummary
                  on:select={() => {
                    dataModelerService.dispatch("setActiveAsset", [EntityType.Model, query.id]);
                  }}
                  on:delete={() => {
                    dataModelerService.dispatch('deleteModel', [query.id]);
                  }}
                  icon={ModelIcon}
                  name={query.name}
                  cardinality={derivedModel?.cardinality ?? 0}
                  profile={derivedModel?.profile ?? []}
                  head={derivedModel?.preview ?? []}
                  sizeInByptes={derivedModel?.sizeInBytes ?? 0}
                  emphasizeTitle ={query?.id === $store?.activeEntity?.id}
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