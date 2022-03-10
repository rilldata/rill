<script>
import { getContext, createEventDispatcher } from "svelte";
import { slide } from "svelte/transition";
import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
import SummaryViewSelector from "$lib/components/column-profile/SummaryViewSelector.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte"
import NavEntry from "$lib/components/column-profile/NavEntry.svelte";
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
            .map(column => ({...column, cardinality: table.cardinality, example: table.head[0] }) ) )?.flat()
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
                                <ContextButton tooltipText="delete" suppressTooltip={true} 
                                    
                                >tt</ContextButton>
                            </span>
                            {/if}
                        </span>
                </div>

        </svelte:fragment>
  </NavEntry>

  {#if show}
    <div  transition:slide|local={{duration: 120 }}>
        <div class="pt-2 pb-2 pl-8">
            <SummaryViewSelector bind:sortMethod bind:previewView />
        </div>
        {#if sortedCategoricals}
            {#each sortedCategoricals as column, i (column.name + column.cardinality )}
                <ColumnProfile
                    bind:this={columns[i]}
                    example={column.example[column.name]}
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
{/if}
</div>