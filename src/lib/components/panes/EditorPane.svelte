<script>
import { getContext } from "svelte";
import { fly } from "svelte/transition";
import { flip } from "svelte/animate";
import Editor from "$lib/components/Editor.svelte";
const store = getContext("rill:app:store");

async function getResultset(query) {
  const response = await fetch("http://localhost:8081/results", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ query }),
  });
  const jsonString = await response.json();
  return JSON.parse(jsonString);
}

export let queryInfo;
export let resultset;
export let query;

</script>

{#if store}
<div class="input-body">
{#each $store.queries as q (q.id)}
<div class="stack" transition:fly={{duration: 100, y: -5}} animate:flip={{duration: 100}}>
  <Editor 
    content={q.query}
    name={q.name}
    on:down={() => { store.moveQueryDown(q.id); }}
    on:up={() => { store.moveQueryUp(q.id); }}
    on:delete={() => { store.deleteQuery(q.id) }}
    on:receive-focus={() => {
        getResultset(q.query).then((returned) => {
          if (returned.queryInfo) {
            queryInfo = returned.queryInfo;
          }
          if (returned.results) {
            resultset = returned.results;
          }
        })
    }}
    on:rename={(evt) => {
      store.changeQueryName(q.id, evt.detail);
    }}
    on:write={async (evt)=> {
        store.editQuery(q.id, evt.detail.content);
        getResultset(evt.detail.content).then((returned) => {
          if (returned.queryInfo) {
            queryInfo = returned.queryInfo;
          }
          if (returned.results) {
            resultset = returned.results;
          }
          if (returned.query) {
            query = returned.query;
          }
        })
    }}
/>
</div>
{/each}
</div>
{/if}

<style>


.input-body {
  padding: 1rem;
}
</style>