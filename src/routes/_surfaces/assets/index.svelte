<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import { flip } from "svelte/animate";

import type { ApplicationStore } from "$lib/application-state-stores/application-store";

import Portal from "$lib/components/Portal.svelte";

import ParquetIcon from "$lib/components/icons/Parquet.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";
import AddIcon from "$lib/components/icons/Add.svelte";
import CollapsibleTableSummary from  "$lib/components/column-profile/CollapsibleTableSummary.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";

import { drag } from '$lib/drag'
import {config, dataModelerService} from "$lib/application-state-stores/application-store";
import type { DerivedTableStore, PersistentTableStore } from "$lib/application-state-stores/table-stores";
import type { DerivedModelStore, PersistentModelStore } from "$lib/application-state-stores/model-stores";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

import { assetVisibilityTween, assetsVisible, layout } from "$lib/application-state-stores/layout-store";
import {uploadTableFiles} from "$lib/util/tableFileUpload";

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

function onTableDrop(e: DragEvent) {
  preventDefault(e);
  if (e.dataTransfer?.files) {
    uploadTableFiles(e.dataTransfer.files, `${config.server.serverUrl}/api`);
  }
}

/** FIXME: what is the correct type for this kind of event? */
function onManualUpload(e: Event) {
  preventDefault(e);
  // @ts-ignore
  if (e?.target?.files) {
    // @ts-ignore
    uploadTableFiles(e.target.files, `${config.server.serverUrl}/api`);
  }
}

function preventDefault(e: DragEvent) {
  e.preventDefault();
  e.stopPropagation();
}

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


    <div class='w-full flex flex-col h-full'>
      <div class='grow' style:outline="1px solid black">
      <header style:height="var(--header-height)" class='sticky top-0 grid align-center bg-white z-50'>
        <h1 class='grid grid-flow-col justify-start gap-x-3 p-4 items-center content-center'>
          <div class='grid  text-white w-5 h-5 items-center justify-center rounded bg-gray-500' style:width="16px" style:height="16px"></div>
          <div class='font-bold'>Rill Developer</div>
        </h1>
      </header>

      <!-- <div style:height="80px"></div> -->

          <div class='pl-4 pb-3 pr-4 pt-5 grid justify-between' style="grid-template-columns: auto max-content;"
               on:drop={onTableDrop}
               on:drag={preventDefault}
               on:dragenter={preventDefault}
               on:dragover={preventDefault}
               on:dragleave={preventDefault}>
            <CollapsibleSectionTitle tooltipText={"tables"} bind:active={showTables}>
              <h4 class='flex flex-row items-center gap-x-2'><ParquetIcon size="16px" /> Tables</h4>

            </CollapsibleSectionTitle>
            
            <ContextButton 
                id={'create-table-button'}
                tooltipText="import csv or parquet file into a table" on:click={async () => {
                  const input = document.createElement('input');
                  input.type = "file";
                  input.multiple = true;
                  input.addEventListener("change", onManualUpload, false);
                  input.click();
                }
              }>
                <AddIcon />
              </ContextButton>
          </div>
            {#if showTables}
              <div class="pb-6"
                   transition:slide|local={{duration:200}}
                   on:drop={onTableDrop}
                   on:drag={preventDefault}
                   on:dragenter={preventDefault}
                   on:dragover={preventDefault}
                   on:dragleave={preventDefault}>
              {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
                <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
                {#each ($persistentTableStore.entities) as { path, tableName, id} (id)}
                  {@const derivedTable = $derivedTableStore.entities.find(t => t["id"] === id)}
                  <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
                    <CollapsibleTableSummary
                      indentLevel={1}
                      icon={ParquetIcon}
                      name={tableName}
                      cardinality={derivedTable?.cardinality ?? 0}
                      profile={derivedTable?.profile ?? []}
                      head={derivedTable?.preview ?? []}
                      {path}
                      sizeInBytes={derivedTable?.sizeInBytes ?? 0}
                      on:delete={() => {
                        dataModelerService.dispatch('dropTable', [tableName]);
                      }}
                    />
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        
          {#if $persistentModelStore && $persistentModelStore.entities}
          <div class='pl-4 pb-3 pr-4 grid justify-between' style="grid-template-columns: auto max-content;"  out:slide={{ duration: 200} }>
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
                  sizeInBytes={derivedModel?.sizeInBytes ?? 0}
                  emphasizeTitle ={query?.id === $store?.activeEntity?.id}
                />
              {/each}
              </div>
            {/if}
          {/if}
      </div>
      <!-- assets pane footer. -->
      <div class='p-3 italic text-gray-800 bg-gray-50 flex items-center text-center justify-center' style:height="var(--header-height)">
        <div class='text-left'>
        Bugs, complaints, feedback? &nbsp;
        </div>
        <a
          target="_blank" 
          class="inline not-italic font-bold text-blue-600 text-right"
          href="http://bit.ly/3jg4IsF"> Ask us on Discord 💬
        </a>
      </div>
    </div>
</div>
