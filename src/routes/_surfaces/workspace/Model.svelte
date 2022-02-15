<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import type { AppStore } from '$lib/app-store';
import { cubicOut as easing } from 'svelte/easing';
import { flip } from "svelte/animate";
import Editor from "$lib/components/Editor.svelte";
import DropZone from "$lib/components/DropZone.svelte";
import {dataModelerService} from "$lib/app-store";

import PreviewTable from "$lib/components/table/PreviewTable.svelte";

const store = getContext("rill:app:store") as AppStore;

let error;

let errorLineNumber;
let errorMessage;

let showPreview = true;

function getErrorLineNumber(errorString) {
  if (!errorString.includes('LINE')) return { message: errorString };
  const [message, linePortion] = errorString.split('LINE ');
  const lineNumber = parseInt(linePortion.split(':')[0]);
  return { message, lineNumber };
};

$: currentQuery = $store?.activeAsset ? $store.models.find(query => query.id === $store.activeAsset.id) : undefined;
</script>

<div class="editor-pane">
  {#if $store && $store.models && currentQuery}
    <div class="input-body p-4 overflow-auto">
    <!-- {#each $store.queries as q, i (q.id)}
    <div class="stack" 
    animate:flip={{duration: 100}}>

      <DropZone 
        padTop={!!$store.queries.length}
        on:source-drop={(evt) => { 
          dataModelerService.dispatch('addQuery', [{ query: evt.detail.props.content, at: i } ]); 
        }} />
        
      <Editor 
        content={q.query}
        name={q.name}
        errorLineNumber={q.id === $store.activeQuery ? errorLineNumber : undefined}
        on:down={() => { dataModelerService.dispatch('moveQueryDown', [{id: q.id}]); }}
        on:up={() => { dataModelerService.dispatch('moveQueryUp', [{id: q.id}]); }}
        on:delete={() => { dataModelerService.dispatch('deleteQuery', [{id: q.id}]); }}
        on:receive-focus={() => {
            dataModelerService.dispatch('setActiveQuery', [{ id: q.id }]);
            dataModelerService.dispatch('updateQueryInformation', [{id: q.id}]);
        }}
        on:release-focus={() => {
          //dataModelerService.dispatch('releaseActiveQueryFocus', [{ id: q.id }]);
        }}
        on:model-profile={() => {
          dataModelerService.dispatch('computeModelProfile', [{ id: q.id }]);
        }}
        on:rename={(evt) => {
          dataModelerService.dispatch('changeQueryName', [{id: q.id, name: evt.detail}]);
        }}
        on:write={(evt)=> {
            dataModelerService.dispatch('setActiveQuery', [{ id: q.id }])
            dataModelerService.dispatch('updateQuery', [{id: q.id, query: evt.detail.content}]);
            dataModelerService.dispatch('updateQueryInformation', [{id: q.id}]);
        }}
    />
    {/key}

    </div>

  {/each} -->

  <div class="stack" >
    {#key currentQuery?.id}
    <Editor 
      content={currentQuery.query}
      name={currentQuery.name}
      errorLineNumber={ currentQuery.id === $store.activeAsset.id ? errorLineNumber : undefined }
      on:down={() => { dataModelerService.dispatch('moveModelDown', [currentQuery.id]); }}
      on:up={() => { dataModelerService.dispatch('moveModelUp', [currentQuery.id]); }}
      on:delete={() => { dataModelerService.dispatch('deleteModel', [currentQuery.id]); }}
      on:receive-focus={() => {
          dataModelerService.dispatch('setActiveAsset', [currentQuery.id, 'model']);
          //dataModelerService.dispatch('updateQueryInformation', [{id: currentQuery.id }]);
      }}
      on:release-focus={() => {
        //dataModelerService.dispatch('releaseActiveQueryFocus', [{ id: q.id }]);
      }}
      on:model-profile={() => {
        //dataModelerService.dispatch('computeModelProfile', [{ id: currentQuery.id }]);
      }}
      on:rename={(evt) => {
        dataModelerService.dispatch('updateModelName', [currentQuery.id, evt.detail]);
      }}
      on:write={(evt)=> {
          dataModelerService.dispatch('setActiveAsset', [currentQuery.id, 'model']);
          dataModelerService.dispatch('updateModelQuery', [currentQuery.id, evt.detail.content ]);
          //dataModelerService.dispatch('updateQueryInformation', [{ id: currentQuery.id }]);
      }}
  />
  {/key}

  </div>
  

  <DropZone end padTop={$store.models.length}
  on:source-drop={(evt) => {
    dataModelerService.dispatch('addModel', [{ query: evt.detail.props.content } ]);
  }} />

</div>

<!-- FIXME: componentize!-->
  {#if $store.activeAsset && $store.models.find(q => q.id === $store.activeAsset.id)?.error}
    <div transition:slide={{ duration: 200, easing }} 
      class="error p-4 m-4 rounded-lg shadow-md"
    >{$store.models.find(q => q.id === $store.activeAsset.id).error}</div>
    {/if}
</div>

<!-- FIXME: componentize!-->

  {/if}
  <div style:font-size="12px" class="overflow-hidden h-full grid p-5" style:grid-template-rows="max-content auto">
    <button on:click={() => { showPreview = !showPreview }}>{#if showPreview}hide{:else}show{/if} preview</button>
    <div class="rounded overflow-auto border border border-gray-300 {!showPreview && 'hidden'}"
    >
      {#if currentQuery?.preview && currentQuery?.profile}
        <PreviewTable rows={currentQuery.preview} columnNames={currentQuery.profile} />
      {:else}
        <div class="p-5  grid items-center justify-center italic">no columns selected</div>
      {/if}
    </div>
  </div>
</div>

<style>

.editor-pane {
  display: grid;
  grid-template-rows: auto 400px;
  /* height: calc(100vh - var(--header-height)); */
  height: calc(100vh - var(--header-height));
}
.error {
  background-color: var(--error-bg);
  color: var(--error-text);
  font-size: 13px;
  align-self: bottom;
}
</style>