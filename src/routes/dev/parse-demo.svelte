<script>
import Editor from "$lib/components/Editor.svelte";
import { queries } from "./_demo-queries"
import { extractCTEs, extractCoreSelectStatements, extractFromStatements, extractJoins, extractSourceTables } from "$lib/util/model-structure";
let whichQuery = 0;

$: content = queries[whichQuery];
let location;

$: ctes = (content?.length) ? extractCTEs(content) : [];
$: selects = (content?.length) ? extractCoreSelectStatements(content || '') : [];
$: fromStatements = (content?.length) ? extractFromStatements(content || '') : [];
$: sourceTables = (content.length) ? extractSourceTables(content || '') : [];

$: joins = (content?.length) ? extractJoins(content || '') : [];
let currentSelection;

const up = () => { 
	whichQuery = Math.min(queries.length-1, whichQuery + 1);
}

const down = () => {
	whichQuery = Math.max(0, whichQuery - 1);
}

$: selections = currentSelection;
</script>

{whichQuery}
<button on:click={down}>prev</button>
<button on:click={up}>next</button>

<div class='grid grid-cols-2'>
{#key queries[whichQuery]}
<Editor 
selections={selections}
content={content}
on:cursor-location={(event) => {
    location = event.detail.location;
    content = event.detail.content;
}}
/>
{/key}

<div>

{#if sourceTables}
<div class="p-3"  on:blur={() => { 
	currentSelection = undefined }} on:mouseout={() => { 
		currentSelection=undefined; 
	}}>
    <div>source tables: {sourceTables.length}</div>
    {#each sourceTables as item}
        <div 
		on:focus={() => { currentSelection = item.tables; }}
		on:mouseover={() => { currentSelection = item.tables; }} 
		class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200  hover:cursor-pointer">
            <b>{item.name}</b>
        </div>
    {/each}
    <!-- {#each sourceJoins as item}
    <div 
    on:focus={() => { currentSelection = item; }}
    on:mouseover={() => { currentSelection = item; }} 
    class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200  hover:cursor-pointer">
        <b>{item.name}</b>
    </div>
{/each} -->
</div>
{/if}

{#if fromStatements}
<div class="p-3"  on:blur={() => { 
	currentSelection = undefined }} on:mouseout={() => { 
		currentSelection=undefined; 
	}}>
    <div>all table refs: {fromStatements.length}</div>
    {#each fromStatements as item}
        <div 
		on:focus={() => { currentSelection = [item]; }}
		on:mouseover={() => { currentSelection = [item]; }} 
		class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200 hover:cursor-pointer">
            <b>{item.name}</b>
        </div>
    {/each}
</div>
{/if}

{#if joins}
<div class="p-3" on:blur={() => { 
	currentSelection = undefined }} on:mouseout={() => { 
		currentSelection=undefined; 
	}}>
    <div>joins: {joins.length}</div>
    {#each joins as item}
        <div
			on:focus={() => { currentSelection = [item]; }}
			on:mouseover={() => { currentSelection = [item]; }} 
			class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200 hover:cursor-pointer">
            <b>{item.name}</b>
        </div>
    {/each}
</div>
{/if}

{#if ctes}
<div class="p-3"  on:blur={() => { 
	currentSelection = undefined }} on:mouseout={() => { 
		currentSelection=undefined; 
	}}>
    <div>ctes: {ctes.length}</div>
    {#each ctes as cte}
        <div 
		on:focus={() => { currentSelection = [cte]; }}
		on:mouseover={() => { currentSelection = [cte]; }} 
		class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200  hover:cursor-pointer">
            <b>{cte.name}</b> = <i>{cte.substring}</i>
        </div>
    {/each}
</div>
{/if}


{#if selects}
<div class="p-3" on:blur={() => { 
	currentSelection = undefined }} on:mouseout={() => { 
		currentSelection=undefined; 
	}}>
    <div>selects: {selects.length}</div>
    {#each selects as select}
        <div
			on:focus={() => { currentSelection = [select]; }}
			on:mouseover={() => { currentSelection = [select]; }} 
			class="text-ellipsis overflow-hidden whitespace-nowrap hover:bg-yellow-200 hover:cursor-pointer">
            <b>{select.name}</b> = <i>{select.expression}</i>
        </div>
    {/each}
</div>
{/if}

</div>

</div>

