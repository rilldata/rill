<script>
import RowTable from "$lib/components/RowTable.svelte";
import ColumnTable from "$lib/components/ColumnTable.svelte";
import RawJSON from "$lib/components/rawJson.svelte";
import RowIcon from "$lib/components/RowIcon.svelte";
import ColumnIcon from "$lib/components/ColumnIcon.svelte";
import JSONIcon from "$lib/components/JsonIcon.svelte";
export let queryInfo;
export let resultset;
export let query;

let outputView = 'row';
let whichTable = {
  row: RowTable,
  column: ColumnTable,
  json: RawJSON
}

$: console.log(query)
</script>



    <div class='source-tables pad-1rem'>
      {#if queryInfo}
      <h3>sources</h3>
      {#each queryInfo as source, i (source.table)}
        <div>
          <h4>{source.table}</h4>
          <table>
          {#each source.info as column}
            <tr>
              <td>
                <div>{column.type.slice(0,1)}</div>
              </td>
              <td>
              <div style="font-weight: semibold;">{column.name} <span style="font-weight: 300; color: #666;">
                  {#if column.pk === 1} (primary){:else}{/if}
                </span></div>

                
              </td>
              <td>

                {`${source.head[0][column.name]}`.slice(0,25)}
              </td>
            </tr>
          {/each}
          </table>
        </div>
      {/each}
      {/if}
    </div>

    {#if resultset}


    <div class='results-container stack-list'>
      <div class="inspector-header pad-1rem">
        <h3>outputs</h3>
        <div class="inspector-button-row">
          <button class='inspector-button' class:selected={outputView === 'row'} on:click={() => { outputView = 'row' }}>
            <RowIcon size={16} />
          </button>
          <button  class='inspector-button'  class:selected={outputView === 'column'} on:click={() => { outputView = 'column' }}>
            <ColumnIcon size={16}  />
          </button>
          <button  class='inspector-button'  class:selected={outputView === 'json'} on:click={() => { outputView = 'json' }}>
            <JSONIcon size={16}  />
          </button>
        </div>
      </div>
      <div class="results pad-1rem" style="padding-top:0px;">
        {#if resultset}
          {#key query}
            <svelte:component this={whichTable[outputView]} data={resultset} />
          {/key}
        {/if}
      </div>
    </div>
    {/if}
    <div>
    </div>

<style>

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
  padding-bottom: 2rem;
}

.source-tables h4 {
  font-weight: black;
  /* border-top: 1px solid #ccc; */
  padding-top: .5rem;
  font-size: 13px;
  margin:0;
  font-weight: 600;
  margin-bottom:.5rem;
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

table tr td:first-child {
  width: 16px;
  color: #aaa;
  border: 2px solid #ccc;
  text-align: center;
  border-radius: .25rem;
  font-size: 10px;
}

table tr td:nth-child(2) {
  padding-left: .5rem;
}

table tr td:last-child {
  text-align: right;
  color: #666;
  font-style: italic;
}

.results {
  overflow: auto;
  max-width: 600px;
}
</style>