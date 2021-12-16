<script>
import { getContext } from "svelte";
import { fly, slide } from "svelte/transition";
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
  return jsonString;
}

async function getDestinationSize(query) {
  const response = await fetch("http://localhost:8081/destination-size", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ query }),
  });
  const jsonString = await response.json();
  return jsonString;
}

// FIXME: this should NOT be props bound elsewhere!
export let queryInfo;
export let resultset;
export let query;
export let destinationInfo;

export let destinationSize;
// let's display the error here.
let error;

let errorLineNumber;
let errorMessage;

let activeQuery;

function getErrorLineNumber(errorString) {
  if (!errorString.includes('LINE')) return { message: errorString };
  const [message, linePortion] = errorString.split('LINE ');
  const lineNumber = parseInt(linePortion.split(':')[0]);
  return { message, lineNumber };
}

function setActiveQuery(qid) {
  activeQuery = qid;
}

function debounce(func, timeout = 300) {
	let timer;
	return (...args) => {
		clearTimeout(timer);
		timer = setTimeout(() => {
			func.apply(this, args);
		}, timeout);
	};
}

const debounceDestinationSize = debounce((q) => {
    getDestinationSize(q).then(returned => {
      destinationSize = returned.size;
  })
}, 1000)

function runAndUpdateInspector(q) {
  destinationSize = undefined;
  debounceDestinationSize(q);
  // getDestinationSize(q).then(returned => {
  //   destinationSize = returned.size;
  // })
  getResultset(q).then((returned) => {
    if (returned.error) {
      error = returned.error;
      const parsedError = getErrorLineNumber(error);
      if (parsedError.message) {
        errorMessage = parsedError.message;
      }
      if (parsedError.lineNumber) {
        errorLineNumber = parsedError.lineNumber;
      }
      return;
    } else {
      error = undefined;
      errorMessage = undefined;
      errorLineNumber = undefined;
    }
    if (returned.queryInfo) {
      queryInfo = returned.queryInfo;
    }
    if (returned.destinationInfo) {
      destinationInfo = returned.destinationInfo;
    }
    if (returned.results) {
      resultset = returned.results;
    }
    if (returned.query) {
      query = returned.query;
    }
  })
}

</script>

<div class=editor-pane>
  {#if store}
  <div class="input-body">
  {#each $store.queries as q (q.id)}
  <div class="stack" transition:fly={{duration: 100, y: -5}} animate:flip={{duration: 100}}>
    <Editor 
      content={q.query}
      name={q.name}
      errorLineNumber={q.id === activeQuery ? errorLineNumber : undefined}
      on:down={() => { store.moveQueryDown(q.id); }}
      on:up={() => { store.moveQueryUp(q.id); }}
      on:delete={() => { store.deleteQuery(q.id) }}
      on:receive-focus={() => {
          setActiveQuery(q.id);
          runAndUpdateInspector(q.query);
      }}
      on:rename={(evt) => {
        store.changeQueryName(q.id, evt.detail);
      }}
      on:write={(evt)=> {
          setActiveQuery(q.id);
          store.editQuery(q.id, evt.detail.content);
          runAndUpdateInspector(evt.detail.content);
      }}
  />
  </div>
  {/each}
  </div>
  {#if error}
    <div transition:slide={{ duration: 100 }} class=error>{error}</div>
    {/if}
  {/if}
</div>

<style>

.editor-pane {
  display: grid;
  --error: 3rem;
  grid-template-rows: auto max-content;
  height: calc(100vh - var(--header-height));
}

.input-body {
  padding: 1rem;
  /* min-height: calc(100vh - var(--header-height) - var(--error)); */
  overflow: auto;
}

.error {
  background-color: var(--error-bg);
  color: var(--error-text);
  /* height: var(--error); */
  font-size: 13px;
  padding: 1rem;
  align-self: bottom;
}
</style>