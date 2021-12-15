<script>
import { slide } from "svelte/transition";
import RowTable from "$lib/components/RowTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import RowIcon from "$lib/components/icons/RowIcon.svelte";
import JSONIcon from "$lib/components/icons/JsonIcon.svelte";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";

import {format} from "d3-format";

const formatCardinality = format(',');
const formatRollupFactor = format(',r')
// FIXME: these shoudl NOT be passed in as props right?
export let queryInfo;
export let destinationInfo;
export let resultset;
export let query;

let outputView = 'row';
let whichTable = {
  row: RowTable,
  json: RawJSON
}

let innerWidth;

function drag(node, params = { minSize: 400, maxSize: 800, property: `--right-sidebar-width` }) {
    let minSize_ = params?.minSize || 400;
    let maxSize_ = params?.maxSize || 800;
    let property_ = params?.property || '--right-sidebar-width';
    let moving = false;
    let xSpace = minSize_;

    node.style.cursor = "move";
    node.style.userSelect = "none";

    function mousedown() {
      moving = true;
    }

    function mousemove(e) {
      if (moving) {
        const size = innerWidth - e.pageX;
        if (size > minSize_ && size < maxSize_) {
          xSpace = size;
        }

        document.body.style.setProperty(property_, `${xSpace}px`)
      }
    }

    function mouseup() {
      moving = false;
    }

    node.addEventListener("mousedown", mousedown);
    window.addEventListener("mousemove", mousemove);
    window.addEventListener("mouseup", mouseup);
    return {
      update() {
        moving = false;
      },
    };
  }

let showSources;
let showOutputs;
let showDestination;
</script>

<svelte:window bind:innerWidth />

<div class='drawer-container'>
  <div class='drawer-handler' use:drag />
  <div class='inspector'>
    {#if destinationInfo && queryInfo}
      <div class='source-tables pad-1rem cost'>
        <div style="font-weight: bold;">
          {formatRollupFactor(queryInfo.reduce((acc,v) => acc + v.cardinality, 0) / destinationInfo.size)}x reduction
        </div>
        <div style="color: #666;">
          {formatCardinality(queryInfo.reduce((acc,v) => acc + v.cardinality, 0))} ⭢
          {formatCardinality(destinationInfo.size)} rows
        </div>
      </div>
    {/if}
    <hr />
    <div class='source-tables pad-1rem'>
      {#if queryInfo}
        <CollapsibleTitle bind:active={showSources}>Sources</CollapsibleTitle>
        {#if showSources}
        <div transition:slide={{duration: 120 }}>
          {#each queryInfo as source, i (source.table)}
            <div>
              <h4>
                <span>
                  {source.table}
                </span>
                <span>
                  {formatCardinality(source.cardinality)} row{#if source.cardinality !== 1}s{/if}
                </span>
              </h4>
              <table>
              {#each source.info as column}
                <tr>
                  <td>
                  <div style="font-weight: semibold;">{column.Field} 
                      <span style="font-weight: 300; font-size:11px; color: #666;">
                        {column.Type}
                      </span>
                      <span style="font-weight: 300; color: #666;">
                      {#if column.pk === 1} (primary){:else}{/if}
                    </span></div> 
                  </td>
                  <td>
                    {(source.head[0][column.Field] !== '' ? `${source.head[0][column.Field]}` : '<empty>').slice(0,25)}
                  </td>
                </tr>
              {/each}
              </table>
            </div>
          {/each}
        </div>
        {/if}
      {/if}
    </div>
    <hr />
    <div class='source-tables pad-1rem'>
      {#if queryInfo}
        <CollapsibleTitle bind:active={showDestination}>Destination</CollapsibleTitle>
        {#if showDestination}
        <div transition:slide={{duration: 120 }}>
              <table>
              {#each destinationInfo.info as column}
                <tr>
                  <!-- <td>
                    <div>{column.Type.slice(0,1)}</div>
                  </td> -->
                  <td>
                  <div style="font-weight: semibold;">{column.Field} 
                    <span style="font-weight: 300; font-size:11px; color: #666;">
                      {column.Type}
                    </span>
                    <span style="font-weight: 300; color: #666;">
                      {#if column.pk === 1} (primary){:else}{/if}
                    </span></div> 
                  </td>
                  <td>
                    {(resultset[0][column.Field] !== '' ? `${resultset[0][column.Field]}` : '<empty>').slice(0,25)}
                  </td>
                </tr>
              {/each}
              </table>
        </div>
        {/if}
      {/if}
    </div>

    {#if resultset}
    <hr />

    <div class='results-container stack-list'>
      <div class="inspector-header pad-1rem">
        <CollapsibleTitle bind:active={showOutputs}>Preview</CollapsibleTitle>
        {#if showOutputs}
        <div class="inspector-button-row">
          <button class='inspector-button' class:selected={outputView === 'row'} on:click={() => { outputView = 'row' }}>
            <RowIcon size={16} />
          </button>
          <button  class='inspector-button'  class:selected={outputView === 'json'} on:click={() => { outputView = 'json' }}>
            <JSONIcon size={16}  />
          </button>
        </div>
        {/if}
      </div>


      {#if showOutputs}
      <div class="results pad-1rem" style="padding-top:0px;">
        {#if resultset}
          {#key query}
            <svelte:component this={whichTable[outputView]} data={resultset} />
          {/key}
        {/if}
      </div>
      {/if}

    </div>
    {/if}
    <div>
    </div>
  </div>
</div>
<style>

hr {
  border: none;
  border-bottom: 1px solid #ddd;
}
.drawer-container {
  display: grid;
  grid-template-columns: max-content auto;
  align-content: stretch;
  align-items: stretch;
}

.cost {
  display: grid;
  grid-template-columns: auto auto;
  justify-content: space-between;
}

.drawer-handler {
  min-width: 1rem;
  position:absolute;
  height: 100%;
  height: calc(100vh - var(--header-height));
  /* background-color: lightgray; */
  transform: translateX(-.5rem);
}

.inspector {
  width: var(--right-sidebar-width, 450px);
  font-size: 13px;

}

.drawer-handler:hover {
  cursor:col-resize;
}

.pad-1rem {
  padding: 1rem;
}

.inspector-header {
  display: grid;
  grid-template-columns: auto max-content;
  align-items: baseline;
  position: sticky;
  top: 0px;
  background-color: white;
}

.inspector-button-row {
  display: grid;
  grid-auto-flow: column;
  justify-content: start;
}

.source-tables {
  display: grid;
  grid-auto-flow: rows;
  grid-gap: 1.25rem;
}

.source-tables h4 {
  font-weight: black;
  /* border-top: 1px solid #ccc; */
  padding-top: .5rem;
  font-size: 13px;
  margin:0;
  font-weight: 600;
  margin-bottom:.5rem;
  display: grid;
  grid-auto-flow: column;
  justify-content: space-between;
}

.source-tables h4 span:nth-child(2) {
  font-weight: normal;
}

h3 {
  margin: 0;
  padding: 0;
  font-size: 13px;
  font-weight: normal;
}

table {
  width: 100%;
  font-size:13px;
  text-align: left;
}

table tr td {
  vertical-align: top;
}

/* table tr td:first-child {
  width: 16px;
  color: #aaa;
  border: 2px solid #ccc;
  text-align: center;
  border-radius: .25rem;
  font-size: 10px;
} */

table tr td:first-child {
  padding-left: .5rem;
}

table tr td:last-child {
  text-align: right;
  color: #666;
  font-style: italic;
}

.results {
  overflow: auto;
  max-width: var(--right-sidebar-width);
}
</style>