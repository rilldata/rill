<script lang="ts">
import { getContext, onMount } from "svelte";
import { tweened } from "svelte/motion"
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { ApplicationStore } from "$lib/app-store";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import AddIcon from "$lib/components/icons/Add.svelte";
import CollapsibleTableSummary from  "$lib/components/column-profile/CollapsibleTableSummary.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

import { drag } from '$lib/drag'
import {dataModelerService} from "$lib/app-store";
import type { DerivedTableStore, PersistentTableStore } from "$lib/tableStores";
import type { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

// FIXME: make thes contextual
import { assetVisibilityTween, assetsVisible } from "$lib/pane-store";

const panes = getContext('rill:app:panes');

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
let width = tweened(400, {duration : 50})

</script>

<div class='flex flex-row-reverse fixed bg-white' style:top="0px" style:height="100vh" style:width="{$panes.left}px">
    <!-- Drawer Handler -->
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize translate-x-2 body-height'
    use:drag={{ side: 'left', minSize: 300, maxSize: 500 }} />

    <div class='assets' bind:this={container} style="width: 100%;">
      <button on:click={() => {
        assetsVisible.set($assetsVisible ? 0 : 1);
      }}>
        gooo
      </button>
      <header class='sticky top-0'>
        <h1  class='grid grid-flow-col justify-start gap-x-3 p-3 items-center content-center'>
          <div class='grid bg-gray-400 text-white w-5 h-5 items-center justify-center rounded'>
            R
          </div>
          <div class='font-normal'>untitled project</div>
        </h1>
      </header>

      <div style:height="80px"></div>

          <div class='pl-3 pb-3 pt-3'>
            <CollapsibleSectionTitle tooltipText={"tables"} bind:active={showTables}>
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
            <CollapsibleSectionTitle  tooltipText={"tables"} bind:active={showModels}>
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


.assets {
  font-size: 12px;
}
</style>