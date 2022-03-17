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
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

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


    <div class='w-full' bind:this={container}>
      <header style:height="var(--header-height)" class='sticky top-0 grid align-center bg-white z-50'>
        <h1 class='grid grid-flow-col justify-start gap-x-3 p-3 items-center content-center'>
          <div class='grid  text-white w-5 h-5 items-center justify-center rounded'>
            <svg width="20" viewBox="0 0 252 254" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M244 242V12C244 9.79086 242.209 8 240 8H12C9.79086 8 8 9.79086 8 12V242C8 244.209 9.79087 246 12 246H240C242.209 246 244 244.209 244 242Z" stroke="black" stroke-width="16"/>
              <path d="M49 46.25V41L47.5098 39.5L43.0392 36.5L38.5686 34.25L28.1373 29H11V43.25L40.8039 46.25L44.5294 50L49 46.25Z" stroke="black" stroke-width="16"/>
              <path d="M109.25 16.0952L119 10L162 8L149 14.5714L135.5 19.1429L124.25 26L104 42L98 29.8095L104 21.4286L109.25 16.0952Z" fill="black" stroke="black" stroke-width="16"/>
              <path d="M195.607 41.2038L172 10H218L238 12L239.5 240.5L179.505 243.887L179.377 244.028H177L179.505 243.887L194.869 226.94L202.246 209.853L211.098 189.793L217 166.019V145.959L215.525 118.47L209.623 85.0376L200.77 53.0909L195.607 41.2038Z" fill="black" stroke="black" stroke-width="16"/>
              <path d="M149.837 201.118L139.858 197.326C139.291 197.11 138.689 197 138.082 197H125.531C125.013 197 124.499 197.08 124.007 197.238L111.13 201.358C110.712 201.492 110.313 201.681 109.944 201.918L102.25 206.88L90.2568 215.459C89.7566 215.816 89.3266 216.264 88.9886 216.777L85.8229 221.589C85.2861 222.405 85 223.361 85 224.337V228.704C85 229.828 85.3789 230.92 86.0756 231.802L86.6109 232.48C87.4946 233.6 88.8142 234.289 90.2378 234.373L99.8555 234.947C100.447 234.982 101.04 234.912 101.607 234.739L108.25 232.72L118.75 228.92L144.217 223.021C144.736 222.901 145.232 222.699 145.687 222.423L151.972 218.601C152.323 218.388 152.646 218.132 152.934 217.841L155.559 215.18C156.482 214.245 157 212.983 157 211.668V210.282C157 209.069 156.559 207.897 155.759 206.985L151.82 202.495C151.282 201.881 150.601 201.408 149.837 201.118Z" stroke="black" stroke-width="16"/>
              <path d="M140.962 55.3574L157.215 54H161.454C161.666 54 161.877 54.0168 162.086 54.0503L165.393 54.5795C166.008 54.6781 166.593 54.9192 167.098 55.2836L170.743 57.9086C171.787 58.6604 172.405 59.8681 172.405 61.1543V63.5975C172.405 64.4694 172.69 65.3173 173.216 66.0123L177.67 71.8935C178.196 72.5886 178.481 73.4365 178.481 74.3083V82.927C178.481 83.5701 178.636 84.2037 178.933 84.7741L179.548 85.9553C179.845 86.5257 180 87.1593 180 87.8024V93.06C180 94.5981 179.118 96 177.731 96.6658L177.458 96.7969C176.206 97.3983 175.355 98.6066 175.211 99.9885L174.857 103.395C174.746 104.46 174.213 105.435 173.378 106.104L166.075 111.948C165.74 112.216 165.364 112.429 164.962 112.577L157.286 115.413C156.27 115.788 155.145 115.735 154.168 115.266L153.479 114.935C152.939 114.676 152.347 114.541 151.748 114.541H145.914C144.881 114.541 143.888 114.142 143.143 113.426L141.913 112.245C140.629 111.012 138.686 110.776 137.145 111.664L134.39 113.251C133.028 114.036 131.332 113.95 130.056 113.031L129.114 112.353L126.835 110.165L123.508 106.969C122.722 106.215 122.278 105.173 122.278 104.084V100.682V97.0298C122.278 96.1013 121.9 95.2129 121.23 94.5698C120.456 93.826 120.077 92.7609 120.207 91.6951L121.519 80.9882V77.2597C121.519 75.941 122.169 74.7071 123.256 73.9612L125.905 72.1441C126.513 71.7268 126.995 71.1493 127.295 70.4756L128.889 66.9051C129.038 66.57 129.233 66.257 129.468 65.975L134.931 59.4163C135.103 59.2099 135.296 59.0213 135.505 58.8534L138.795 56.2205C139.416 55.724 140.17 55.4235 140.962 55.3574Z" stroke="black" stroke-width="16"/>
              <path d="M12.0695 80.7414L10.5957 81.4723C9.9316 81.8017 9.50414 82.4713 9.48499 83.2124L8.03299 139.382C8.01287 140.16 8.44622 140.879 9.1437 141.225L9.51613 141.41L11.0323 142.162C11.5376 142.413 12.5484 143.064 12.5484 143.666C12.5484 144.418 14.8226 143.666 15.5806 143.666H17.6587C17.7894 143.666 17.9194 143.653 18.0476 143.628L20.6878 143.104C21.292 142.984 21.8019 142.589 22.1219 142.062C22.4225 141.568 22.8008 141.016 23.1613 140.658C23.4333 140.388 23.6139 140.007 23.7302 139.624C23.951 138.897 24.3453 138.174 25.0605 137.917C25.4767 137.768 25.9002 137.65 26.1935 137.65C26.6017 137.65 27.0328 137.877 27.3481 138.102C27.6 138.282 27.8619 138.454 28.1573 138.546C28.9961 138.807 30.2369 139.154 30.7419 139.154H34.5323C35.3773 139.154 36.0262 138.911 36.4274 138.679C36.6952 138.523 36.9593 138.352 37.2531 138.255L41.0481 137C41.2513 136.933 41.4423 136.833 41.614 136.706L43.8457 135.045C44.1969 134.784 44.4527 134.415 44.5739 133.994L45.6593 130.226C45.8147 129.686 46.1897 129.237 46.6926 128.988L46.9994 128.836C47.2756 128.699 47.5849 128.636 47.8777 128.539C48.1813 128.439 48.5677 128.24 48.9355 127.875C49.6935 127.123 51.2097 127.875 51.9677 127.875C52.5045 127.875 53.4767 126.697 54.0435 125.905C54.1776 125.718 54.2707 125.505 54.3271 125.282L54.9394 122.852C54.9796 122.693 55 122.528 55 122.364V117.154C55 116.789 54.9002 116.431 54.7113 116.119L53.0145 113.314C52.8257 113.002 52.7258 112.644 52.7258 112.279V107.573V104.289C52.7258 103.977 52.7991 103.668 52.9399 103.389L54.4733 100.347C54.7987 99.7015 54.7505 98.9308 54.3473 98.331L52.1755 95.0996C52.0382 94.8953 51.8643 94.7181 51.6627 94.5769L39.3388 85.9481C39.1677 85.8283 39.0165 85.6825 38.8906 85.516L35.1324 80.5457C34.7544 80.0457 34.1639 79.7519 33.5371 79.7519H32.2581L27.9353 79.1394C27.3156 79.0516 26.6906 79.2589 26.2462 79.6997L25.3143 80.6241C24.9092 81.0259 24.3516 81.2356 23.7821 81.2003L13.0819 80.5369C12.7324 80.5153 12.3833 80.5858 12.0695 80.7414Z" stroke="black" stroke-width="16"/>
              </svg>
              
          </div>
          <div class='font-bold'>Data Modeler Prototype</div>
        </h1>
      </header>

      <!-- <div style:height="80px"></div> -->

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