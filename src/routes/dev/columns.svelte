<script>
import { getContext, createEventDispatcher } from "svelte";
import CategoricalEntry from "$lib/components/collapsible-table-summary/CategoricalEntry.svelte";
import SummaryViewSelector from "$lib/components/collapsible-table-summary/SummaryViewSelector.svelte";
import ContextButton from "$lib/components/collapsible-table-summary/ContextButton.svelte"
import NavEntry from "$lib/components/collapsible-table-summary/NavEntry.svelte";
import icon from "$lib/components/icons/List.svelte"
const store = getContext("rill:app:store");

const dispatch = createEventDispatcher();


let width = 450;

let columns = [];

let sortMethod;
let previewView;

let name = "cool.sql"
let titleElementHovered = false;
let show = true;
let emphasizeTitle = true;

$: categoricals = $store?.tables
    ?.map(table => 
        table.profile
            //.filter(column => CATEGORICALS.has(column.type))
            .map(column => ({...column, cardinality: table.cardinality }) ) )?.flat()
let sortedCategoricals = categoricals;
$: sortedCategoricals = sortMethod !== undefined && categoricals !== undefined ? categoricals.sort(sortMethod) : undefined;

// threshold to remove all summary items.
$: hideRight = width < 250;
// threshold to remove null percentages.
$: hideNullPercentage = width < 400;



</script>

<input type=range min=200 max=900 bind:value={width} />

<button on:click={() => { columns.forEach(col => { col.close() } ) }}>close all</button>

{width}


<div style:font-size=12px style:width="{width}px" style:height="200vh">
    <NavEntry
    expanded={show}
    selected={emphasizeTitle}
    bind:hovered={titleElementHovered}
    on:select-body={() => { 
        dispatch('select');
    }}
    on:expand={() => {
        show = !show;
        dispatch('expand');
    }}
    {icon} >
    <span class:font-bold={emphasizeTitle}>
        {#if name.split('.').length > 1}
            {name.split('.').slice(0, -1).join('.')}<span class='text-gray-500 italic pl-1'>.{name.split('.').slice(-1).join('.')}</span>
        {:else}
            {name}
        {/if}
    </span>
        <svelte:fragment slot='contextual-information'>
                <div class='italic text-gray-600'>

                        <span class="grid grid-flow-col gap-x-2 text-gray-500 text-clip overflow-hidden whitespace-nowrap ">
                            {#if titleElementHovered || emphasizeTitle}
                            <!-- <span ><span>{cardinality !== undefined && cardinality !== NaN ? formatInteger(interimCardinality) : "no"}</span> row{#if cardinality !== 1}s{/if}</span> -->
                            <span class='self-center'>
                                <ContextButton tooltipText="delete, more..." suppressTooltip={true} 
                                    
                                >tt</ContextButton>
                            </span>
                            {/if}
                        </span>
                </div>

        </svelte:fragment>
  </NavEntry>


<SummaryViewSelector bind:sortMethod bind:previewView />

{#if sortedCategoricals}
    {#each sortedCategoricals as column, i (column.name + column.cardinality )}
        <CategoricalEntry
            bind:this={columns[i]}
            
            containerWidth={width}

            {hideRight}
            {hideNullPercentage}
            hideSummaryPreview={false}

            view={previewView}

            name={column.name}
            type={column.type}
            summary={column.summary}
            totalRows={column.cardinality}
            nullCount={column.nullCount}
        />
    {/each}
{/if}
</div>