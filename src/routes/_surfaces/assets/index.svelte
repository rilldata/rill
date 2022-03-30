<script lang="ts">
import { getContext, onMount } from "svelte";
import { tweened } from "svelte/motion"
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { ApplicationStore } from "$lib/app-store";

import Portal from "$lib/components/Portal.svelte";

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
import { EntityStatus, EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

import { assetVisibilityTween, assetsVisible, layout } from "$lib/layout-store";

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

let view = 'assets';

</script>

<div class='
  border-r 
  border-transparent 
  fixed 
  overflow-auto 
  hover:border-gray-200 
  transition-colors
  h-screen
  bg-white
' 
  class:hidden={$assetVisibilityTween === 1}
  class:pointer-events-none={!$assetsVisible}
  style:top="0px" style:width="{$layout.assetsWidth}px">
    
    <!-- draw handler -->
    {#if $assetsVisible}
      <Portal>
        <div 
        class='fixed z-50 drawer-handler w-4 hover:cursor-col-resize -translate-x-2 h-screen'
        style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth}px"
        use:drag={{ minSize: 300, maxSize:500,  side: 'assetsWidth',  }} />
      </Portal>
    {/if}


    <div class='w-full'>
      <header style:height="var(--header-height)" class='sticky top-0 grid align-center bg-white z-50'>
        <h1 class='grid grid-flow-col justify-start gap-x-3 p-4 items-center content-center'>
          <div class='grid  text-white w-5 h-5 items-center justify-center rounded bg-gray-500' style:width="16px" style:height="16px"></div>
          <div class='font-bold'>Rill Developer</div>
        </h1>
      </header>

      <!-- <div style:height="80px"></div> -->

          <div class='pl-4 pb-3 pt-5'>
            <CollapsibleSectionTitle tooltipText={"tables"} bind:active={showTables}>
              <h4 class='flex flex-row items-center gap-x-2'><ParquetIcon size="16px" /> Tables</h4>

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
                      indentLevel={1}
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
          <div class='pl-4 pb-3 pr-4 grid justify-between' style="grid-template-columns: auto max-content;">
            <CollapsibleSectionTitle  tooltipText={"tables"} bind:active={showModels}>
                <h4 class='flex flex-row items-center gap-x-2'><ModelIcon size="16px" /> Models</h4>
              </CollapsibleSectionTitle>
              <ContextButton 
                id={'create-model-button'}
                tooltipText="create a new model" on:click={async () => {
                // create the new model.
                let response = await dataModelerService.dispatch("addModel", [{}]);
                // change the active asset to the new model.
                dataModelerService.dispatch("setActiveAsset", [EntityType.Model, response.id]);
                // if the models are not visible in the assets list, show them.
                if (!showModels) {
                  showModels = true;
                }
              }}>
                <AddIcon />
              </ContextButton>

            </div>
            {#if showModels}
              <div class='pb-6 justify-self-end'  transition:slide={{duration:200}} id="assets-model-list">
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
                  indentLevel={1}
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