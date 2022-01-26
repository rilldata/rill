<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import type { AppStore } from '$lib/app-store';
import { cubicOut as easing } from 'svelte/easing';
import { flip } from "svelte/animate";
import Editor from "$lib/components/Editor.svelte";
import DropZone from "$lib/components/DropZone.svelte";

const store = getContext("rill:app:store") as AppStore;

let error;

let errorLineNumber;
let errorMessage;

function getErrorLineNumber(errorString) {
  if (!errorString.includes('LINE')) return { message: errorString };
  const [message, linePortion] = errorString.split('LINE ');
  const lineNumber = parseInt(linePortion.split(':')[0]);
  return { message, lineNumber };
};

$: currentQuery = $store?.activeAsset ? $store.queries.find(query => query.id === $store.activeAsset.id) : undefined;
</script>

<div class="editor-pane">
  {#if $store && $store.queries && currentQuery}
    <div class="input-body p-4 overflow-auto">
    <!-- {#each $store.queries as q, i (q.id)}
    <div class="stack" 
    animate:flip={{duration: 100}}>

      <DropZone 
        padTop={!!$store.queries.length}
        on:source-drop={(evt) => { 
          store.action('addQuery', { query: evt.detail.props.content, at: i } ); 
        }} />
        
      <Editor 
        content={q.query}
        name={q.name}
        errorLineNumber={q.id === $store.activeQuery ? errorLineNumber : undefined}
        on:down={() => { store.action('moveQueryDown', {id: q.id}); }}
        on:up={() => { store.action('moveQueryUp', {id: q.id}); }}
        on:delete={() => { store.action('deleteQuery', {id: q.id}); }}
        on:receive-focus={() => {
            store.action('setActiveQuery', { id: q.id });
            store.action("updateQueryInformation", {id: q.id});
        }}
        on:release-focus={() => {
          //store.action('releaseActiveQueryFocus', { id: q.id });
        }}
        on:model-profile={() => {
          store.action('computeModelProfile', { id: q.id });
        }}
        on:rename={(evt) => {
          store.action('changeQueryName', {id: q.id, name: evt.detail});
        }}
        on:write={(evt)=> {
            store.action('setActiveQuery', { id: q.id })
            store.action("updateQuery", {id: q.id, query: evt.detail.content});
            store.action("updateQueryInformation", {id: q.id});
        }}
    />

    </div>

  {/each} -->

  <div class="stack" >
    {#key currentQuery?.id}
    <Editor 
      content={currentQuery.query}
      name={currentQuery.name}
      errorLineNumber={ currentQuery.id === $store.activeAsset.id ? errorLineNumber : undefined }
      on:down={() => { store.action('moveQueryDown', {id: currentQuery.id}); }}
      on:up={() => { store.action('moveQueryUp', {id: currentQuery.id}); }}
      on:delete={() => { store.action('deleteQuery', {id: currentQuery.id}); }}
      on:receive-focus={() => {
          store.action('setActiveAsset', { id: currentQuery.id, assetType: 'model' });
          store.action("updateQueryInformation", {id: currentQuery.id });
      }}
      on:release-focus={() => {
        //store.action('releaseActiveQueryFocus', { id: q.id });
      }}
      on:model-profile={() => {
        store.action('computeModelProfile', { id: currentQuery.id });
      }}
      on:rename={(evt) => {
        store.action('changeQueryName', {id: currentQuery.id, name: evt.detail });
      }}
      on:write={(evt)=> {
          store.action('setActiveAsset', { id: currentQuery.id, assetType: 'model' });
          store.action("updateQuery", { id: currentQuery.id, query: evt.detail.content });
          store.action("updateQueryInformation", { id: currentQuery.id });
      }}
  />
  {/key}

  </div>
  

  <DropZone end padTop={$store.queries.length}
  on:source-drop={(evt) => {
    store.action('addQuery', { query: evt.detail.props.content } );
  }} />

</div>

<!-- FIXME: componentize!-->
  {#if $store.activeAsset && $store.queries.find(q => q.id === $store.activeAsset.id)?.error}
    <div transition:slide={{ duration: 200, easing }} 
      class="error p-4 m-4 rounded-lg shadow-md"
    >{$store.queries.find(q => q.id === $store.activeAsset.id).error}</div>
    {/if}
  {/if}
</div>

<style>

.editor-pane {
  display: grid;
  grid-template-rows: auto max-content;
  height: calc(100vh - var(--header-height));
}

.error {
  background-color: var(--error-bg);
  color: var(--error-text);
  font-size: 13px;
  align-self: bottom;
}
</style>