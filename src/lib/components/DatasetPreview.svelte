<script lang="ts">
import { onMount } from "svelte";
import { slide, fade, fly } from "svelte/transition";
import { tweened } from "svelte/motion";
import { cubicInOut as easing } from "svelte/easing";
import { format } from "d3-format";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";
import TopKSummary from "$lib/components/TopKSummary.svelte";
import IconButton from "$lib/components/IconButton.svelte";

import BarAndLabel from "$lib/components/BarAndLabel.svelte";

import { dropStore } from '$lib/drop-store';
import type { SvelteComponent } from "svelte/internal";

import { intervalToTimestring, formatCardinality } from "../../util/formatters";

export let icon:SvelteComponent;
export let name:string;
export let path:string;
export let cardinality:number;
export let profile:any;
export let head:any; // FIXME
export let sizeInBytes:number;
export let categoricalSummaries:any; // FIXME
export let timestampSummaries:any; // FIXME
export let numericalSummaries:any; // FIXME
export let nullCounts:any;
export let collapseWidth = 200 + 120 + 16;
export let emphasizeTitle:boolean = false;
export let draggable = true;

let colSizer;

const formatInteger = format(',');
const percentage = format('.1%');


let container;

let containerWidth = 0;
let show = false;

function humanFileSize(size:number) {
    var i = Math.floor( Math.log(size) / Math.log(1024) );
    return ( size / Math.pow(1024, i) ).toFixed(2) + ['B', 'K', 'M', 'G'][i];
};

onMount(() => {
    const observer = new ResizeObserver(entries => {
        entries.forEach((entry) => {
            containerWidth = container.clientWidth;
        });
    });
    observer.observe(container);
});

$: collapseGrid = containerWidth < collapseWidth;

let cardinalityTween = tweened(cardinality, { duration: 600, easing });
let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });
$: cardinalityTween.set(cardinality);
$: sizeTween.set(sizeInBytes);

let selectingColumns = false;
let selectedColumns = [];

let showSummaries = true;
let summaryColumns = [];

// FIXME: we should be sorting columns by type (timestamp, numeric, categorical)
// on the backend, not here.

function sortByCardinality(a,b) {
    if (categoricalSummaries) {
        if (categoricalSummaries[a.name]?.cardinality < categoricalSummaries[b.name]?.cardinality) {
            return 1;
        } else if (categoricalSummaries[a.name]?.cardinality > categoricalSummaries[b.name]?.cardinality) {
            return -1;
        } else {
            return sortByName(a,b);
        }
    } else {
        return 0;
    }
}

function sortByNullity(a,b) {
    if (nullCounts) {
        if (nullCounts[a.name] < nullCounts[b.name]) {
            return 1;
        } else if ((nullCounts[a.name] > nullCounts[b.name])) {
            return -1;
        } else {
            return sortByName(a,b);
        }
    } else {
        return 0;
    }
}

function sortByName(a,b) {
    return (a.name > b.name) ? 1 : -1;
}

function defaultSort(a, b) {
    if (a.type === 'BYTE_ARRAY' &&  b.type !== 'BYTE_ARRAY') return 1;
    if (a.type !== 'BYTE_ARRAY' &&  b.type === 'BYTE_ARRAY') return -1;
    if (a.type !== 'BYTE_ARRAY' && b.type !== 'BYTE_ARRAY') return sortByNullity(b,a);
    return sortByCardinality(a,b);
}

let sortMethod = defaultSort;
$: sortedProfile = [...profile].sort(sortMethod);

let previewView = 'card';
</script>

<div bind:this={container}>
    <div {draggable} 
        class="active:cursor-grabbing"
        on:dragstart={(evt) => {
            var elem = document.createElement("div");
            elem.id = "drag-ghost";
            elem.textContent = `${name}`;
            elem.style.position = "absolute";
            elem.style.top = "-1000px";
            elem.style.fontSize = '12px';
            elem.style.transform = 'translateY(-5em)';
            elem.classList.add('draggable');
            document.body.appendChild(elem);
            evt.dataTransfer.setDragImage(elem, 0, 0);
            dropStore.set({
                type: "source-to-query",
                props: {
                    content: `SELECT \n  ${selectingColumns && selectedColumns.length ? selectedColumns.join(',\n  ') : '*' }\nFROM '${path}';`,
                    name: 'whatever.sql'
                }
            });
        }}
        on:dragend={() => {
            var ghost = document.getElementById("drag-ghost");
            if (ghost.parentNode) {
                ghost.parentNode.removeChild(ghost);
            }
            dropStore.set(undefined);
        }}>
    <CollapsibleTitle {icon} bind:active={show}>
        <span class:font-bold={emphasizeTitle} class:italic={selectingColumns}>
            {name}{#if selectingColumns}&nbsp;<span class="font-bold"> *</span>{/if}
        </span>
            <svelte:fragment slot='contextual-information'>
                    <div class='italic text-gray-600'>
                        {#if selectingColumns}
                            {#if selectedColumns.length}
                                selected {selectedColumns.length} column{#if selectedColumns.length > 1}s{/if}
                            {:else}
                                select columns
                            {/if}
                        {:else}
                            {formatInteger($cardinalityTween)} row{#if cardinality !== 1}s{/if}{#if !collapseGrid && sizeInBytes !== undefined}, {humanFileSize($sizeTween)}{/if}
                        {/if}
                    </div>

            </svelte:fragment>
      </CollapsibleTitle>
    </div>
    {#if show}
        <div class="pt-1 pl-accordion" transition:slide|local={{duration: 120 }}>

            <div class:hidden={collapseGrid} class='grid grid-flow-col justify-between'>
                <div class='grid justify-start grid-cols-3 items-baseline pb-2 pt-2'>
                    <IconButton title="sort by cardinality" selected={sortMethod === defaultSort} on:click={() => {
                        sortMethod = defaultSort;
                    }}>||</IconButton>

                    <IconButton 
                        title="sort by null % of field"
                        selected={sortMethod === sortByNullity}  on:click={() => {
                        sortMethod = sortByNullity;
                    }}>∅</IconButton>
                    <IconButton title="sort by name" selected={sortMethod === sortByName}  on:click={() => {
                        sortMethod = sortByName;
                    }}>AZ</IconButton>
                </div>

                <div class='grid justify-start grid-cols-2 items-baseline pb-2 pt-2'>
                    {#if draggable}
                    <IconButton 
                        title="select columns"
                        selected={selectingColumns} on:click={() => {
                        selectingColumns = !selectingColumns;
                    }}>*</IconButton>
                    {/if}
                    <IconButton title="summarize columns" selected={showSummaries} on:click={() => {
                        showSummaries = !showSummaries;
                    }}>
                        &
                    </IconButton>
                </div>

                <div class='grid justify-end grid-flow-col pb-2 pt-2 '>
                    <IconButton 
                        title="show summaries (cardinalities, null %, and time range)"
                        selected={previewView === 'card'} 
                        on:click={() => {
                            previewView = 'card';
                    }}>||</IconButton>
                    <IconButton 
                        title="show example value"
                        selected={previewView === 'example'}  
                        on:click={() => {
                            previewView = 'example';
                    }}>Ex</IconButton>
                </div>
            </div>
            <div 
                class="gap-x-4 w-full" 
                class:grid={!collapseGrid} 
                class:block={collapseGrid}
                style="grid-template-columns: minmax(80px, auto) {previewView === 'example' ? "minmax(108px, 164px)" : "auto 68px"};"
            >
                {#if sortedProfile}
                    {#each sortedProfile as column}
                    <div 
                        class="font-medium break-word grid gap-x-2 items-baseline" 
                        style="grid-template-columns: minmax(90px, max-content) auto;" 
                        bind:this={colSizer}>
                        <button
                            title={column.name}
                            class='text-ellipsis overflow-hidden whitespace-nowrap break-all text-left {(selectingColumns || showSummaries) ? 'hover:underline' : ''}' 
                            class:font-bold={
                                (selectingColumns && selectedColumns.includes(column.name)) ||
                                (showSummaries && summaryColumns.includes(column.name))
                            } 
                            on:click={() => {
                            if (selectingColumns) {
                                if (selectedColumns.includes(column.name)) {
                                    selectedColumns = selectedColumns.filter(c => c !== column.name)
                                } else {
                                    selectedColumns = [...selectedColumns, column.name];
                                }
                            } else if (showSummaries) {
                                if (summaryColumns.includes(column.name)) {
                                    summaryColumns = summaryColumns.filter(c => c !== column.name)
                                } else {
                                    summaryColumns = [...summaryColumns, column.name];
                                }
                            }
                        }}>
                            {column.name} 
                        </button>
                        <div
                        title={column.type} 
                        class="text-ellipsis overflow-hidden whitespace-nowrap " style="font-size:11px; color: #959595;">
                            {column.type}
                            <!-- <span class='italic text-gray-600' style="font-size:10px;">{(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>').slice(0,25)}</span> -->
                            
                        </div>
                    </div>
                    {#if previewView === 'card'}
                    <div 
                        class='justify-self-end text-right text-gray-500  break-all' class:hidden={collapseGrid}>
                        {#if categoricalSummaries && column.name in categoricalSummaries}
                            <!-- <span title="{column.name}: {formatCardinality(categoricalSummaries[column.name].cardinality)} unique values"> -->
                            <BarAndLabel value={categoricalSummaries[column.name].cardinality / cardinality}>
                                |{formatInteger(categoricalSummaries[column.name].cardinality)}|
                            <!-- </span> -->
                            </BarAndLabel>
                        {/if}
                        {#if timestampSummaries && column.name in timestampSummaries}
                            {intervalToTimestring(timestampSummaries[column.name].interval)} 
                        {/if}
           
                        <!-- {(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>').slice(0,25)} -->
                    </div>
                    <div class='self-stretch text-right text-gray-500 break-all' class:hidden={collapseGrid}>
                        <!-- && !(summaryColumns.includes(column.name) && 
                            (categoricalSummaries && column.name in categoricalSummaries && (categoricalSummaries[column.name].topK.slice(0,5).map(k => k.value).includes(null)) )) -->

                            <!-- in:fly={{ duration: 100, x: -10, y: 20, delay: 100 }} out:fly={{ duration: 50, x: -10, y: 20 }} -->
                        {#if nullCounts && column.name in nullCounts && cardinality}
                            <div  >
                                <BarAndLabel
                                title="{column.name}: {percentage(nullCounts[column.name] / cardinality)} of the values are null"
                                bgColor={nullCounts[column.name] === 0 ? 'bg-white' : 'bg-gray-50'}
                                value={nullCounts[column.name] / cardinality || 0}>
                                            <span class:text-gray-300={nullCounts[column.name] === 0}>∅ {percentage(nullCounts[column.name] / cardinality)}</span>
                                </BarAndLabel>
                            </div>
                        {/if}
                    </div>
                    {:else}
                        <div class='text-gray-500 italic text-right text-ellipsis overflow-hidden whitespace-nowrap ' class:hidden={collapseGrid}>
                            {(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>')}
                        </div>
                    {/if}

                    <!-- 
                        summary element.
                        For topK, the labels should be firstColWidth wide.
                    -->
                    <!-- categorical summaries -->
                    {#if showSummaries && summaryColumns.includes(column.name) && categoricalSummaries && column.name in categoricalSummaries}
                    <div style="grid-column: 1 / -1;" class='pt-3 pb-3' transition:slide={{duration: 200 }}>
                        <TopKSummary 
                            totalRows={cardinality}
                            topK={categoricalSummaries[column.name].topK}
                            displaySize={collapseGrid ? 'sm' : 'md'}
                        />
                    </div>
                    {/if}

                    <!-- timestamp summaries -->
                    {#if showSummaries && summaryColumns.includes(column.name) && timestampSummaries && column.name in timestampSummaries}
                        <div  style="grid-column: 1 / -1;" class='pt-3 pb-3 pl-3 pr-3' transition:slide={{duration: 200 }}>

                                {intervalToTimestring(timestampSummaries[column.name].interval)} 
                                ({timestampSummaries[column.name].min}
                                    to
                                    {timestampSummaries[column.name].max}
                                )
                        </div>
                    {/if}

                    <!-- numerical summaries -->
                        {#if showSummaries && summaryColumns.includes(column.name) && numericalSummaries && column.name in numericalSummaries}
                        <div  style="grid-column: 1 / -1;" class='pt-3 pb-3 pl-3 pr-3' transition:slide={{duration: 200 }}>
                            <pre>
                                {JSON.stringify(numericalSummaries[column.name], null, 2)}
                            </pre>
                        </div>
                    {/if}

                    {/each}
                {/if}
            </div>
        </div>
    {/if}
  </div>