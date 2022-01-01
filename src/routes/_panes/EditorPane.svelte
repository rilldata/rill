<script>
import { getContext } from "svelte";
import { fly, slide } from "svelte/transition";
import { cubicOut as easing } from 'svelte/easing';
import { flip } from "svelte/animate";
import Editor from "$lib/components/Editor.svelte";
import DropZone from "$lib/components/DropZone.svelte";
const store = getContext("rill:app:store");

let error;

let errorLineNumber;
let errorMessage;

function getErrorLineNumber(errorString) {
  if (!errorString.includes('LINE')) return { message: errorString };
  const [message, linePortion] = errorString.split('LINE ');
  const lineNumber = parseInt(linePortion.split(':')[0]);
  return { message, lineNumber };
}
$: console.log($store?.queries?.length)
</script>

<div class="editor-pane">
  {#if $store && $store.queries}
    <div class="input-body p-4 overflow-auto">
    {#each $store.queries as q, i (q.id)}
    <div class="stack" 
    animate:flip={{duration: 100}}>

      <DropZone 
        padTop={$store.queries.length}
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

  {/each}

  <DropZone end         padTop={$store.queries.length}
  on:source-drop={(evt) => {
    store.action('addQuery', { query: evt.detail.props.content } );
  }} />

</div>

  {#if $store.activeQuery && $store.queries.find(q => q.id === $store.activeQuery)?.error}
    <div transition:slide={{ duration: 200, easing }} 
      class="error p-4 m-4 rounded-lg shadow-md"
    >{$store.queries.find(q => q.id === $store.activeQuery).error}</div>
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