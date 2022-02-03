<script lang="ts">
import { onMount } from "svelte";
import { fade, slide, fly } from "svelte/transition";
import { tweened } from "svelte/motion";
import { cubicInOut as easing, cubicOut } from "svelte/easing";
import { format } from "d3-format";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";
import TopKSummary from "$lib/components/TopKSummary.svelte";
import IconButton from "$lib/components/IconButton.svelte";

import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import Spinner from "$lib/components/Spinner.svelte";

import { dropStore } from '$lib/drop-store';
import type { SvelteComponent } from "svelte/internal";
import Histogram from "$lib/components/Histogram.svelte";
import SummaryAndHistogram from "$lib/components/SummaryAndHistogram.svelte";

import { horizontalSlide } from "$lib/transitions"

import { intervalToTimestring, formatCardinality } from "$lib/util/formatters";

export let icon:SvelteComponent;
export let name:string;
export let path:string;
export let cardinality:number;
export let profile:any;
export let head:any; // FIXME
export let sizeInBytes:number;
export let collapseWidth = 200 + 120 + 16;
export let emphasizeTitle:boolean = false;
export let draggable = true;

let colSizer;

const formatInteger = format(',');
const percentage = format('.1%');

let containerWidth = 0;
export let show = false;

function humanFileSize(size:number) {
    var i = Math.floor( Math.log(size) / Math.log(1024) );
    return ( size / Math.pow(1024, i) ).toFixed(2) + ['B', 'K', 'M', 'G'][i];
};

let container;
onMount(() => {
    const observer = new ResizeObserver(entries => {
        containerWidth = container.clientWidth;
    });
    observer.observe(container);
});

$: collapseGrid = containerWidth < collapseWidth;

let cardinalityTween = tweened(cardinality, { duration: 600, easing });
let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });

$: cardinalityTween.set(cardinality || 0);
$: interimCardinality = ~~$cardinalityTween;
$: sizeTween.set(sizeInBytes || 0);

let selectingColumns = false;
let selectedColumns = [];

let showSummaries = true;
let summaryColumns = [];

// FIXME: we should be sorting columns by type (timestamp, numeric, categorical)
// on the backend, not here.

function sortByCardinality(a,b) {
    if (a.summary && b.summary) {
        if (a.summary?.cardinality < b.summary?.cardinality) {
            return 1;
        } else if (a.summary?.cardinality > b.summary?.cardinality) {
            return -1;
        } else {
            return sortByName(a,b);
        }
    } else {
        return 0;
    }
}

function sortByNullity(a,b) {
    if (a.nullCount !== undefined && b.nullCount !== undefined) {
        if (a.nullCount < b.nullCount) {
            return 1;
        } else if ((a.nullCount > b.nullCount)) {
            return -1;
        } else {
            const byType = sortByType(a,b);
            if (byType) return byType;
            return sortByName(a,b);
        }
    }

    return sortByName(a,b);
}

function sortByType(a,b) {
    if (categoricals.has(a.type) && !categoricals.has(b.type)) return 1;
    if (!categoricals.has(a.type) && categoricals.has(b.type)) return -1;
    if ((a.conceptualType === 'TIMESTAMP' || a.type === 'TIMESTAMP') && (b.conceptualType !== 'TIMESTAMP' && b.type !== 'TIMESTAMP')) {
                return -1;
    } else if ((a.conceptualType !== 'TIMESTAMP' && a.type !== 'TIMESTAMP') && (b.conceptualType === 'TIMESTAMP' || b.type ==='TIMESTAMP')) {
        return 1;
    }
    return 0;
}

function sortByName(a,b) {
    return (a.name > b.name) ? 1 : -1;
}

const categoricals = new Set(['BYTE_ARRAY', 'VARCHAR'])

function defaultSort(a, b) {
    const byType = sortByType(a,b);
    if (byType !== 0) return byType;
    if (categoricals.has(a.type) && !categoricals.has(b.type)) return sortByNullity(b,a);
    return sortByCardinality(a,b);
}

let sortMethod = defaultSort;
$: sortedProfile = [...profile].sort(sortMethod);

let previewView = 'card';

// FIXME: replace once we work through logical types in parquet_scan
function typeToSymbol(fieldType) {
    if (fieldType === 'BYTE_ARRAY' || fieldType === 'VARCHAR') {
        return {symbol: "C", text: 'categorical (BYTE_ARRAY)', color: 'red'};
    } else {
        return {symbol: fieldType.slice(0,1), text: fieldType, color: 'sky'};
    }
}

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
                            <span>
                                {#if selectedColumns.length}
                                    selected {selectedColumns.length} column{#if selectedColumns.length > 1}s{/if}
                                {:else}
                                    select columns
                                {/if}
                            </span>
                        {:else}
                            <span class="grid grid-flow-col">
                                <span><span>{cardinality !== undefined && cardinality !== NaN ? formatInteger(interimCardinality) : "no"}</span> row{#if cardinality !== 1}s{/if}</span>{#if !collapseGrid && sizeInBytes !== undefined}<span style="display: inline-block; text-overflow: clip; white-space: nowrap;
                                " transition:horizontalSlide|local={{duration: 250 + sizeInBytes / 5000000 * 1.3}}>, {sizeInBytes !== undefined && $sizeTween !== NaN && sizeInBytes !== NaN ? humanFileSize($sizeTween) : ''}</span>{/if}
                            </span>
                        {/if}
                    </div>

            </svelte:fragment>
      </CollapsibleTitle>
    </div>
    {#if show}
        <div class="pt-1 pl-accordion" transition:slide|local={{duration: 120 }}>
            <div class='grid grid-flow-col justify-between'>
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
            <!--                 class:grid={!collapseGrid} 
                class:block={collapseGrid} -->
            <div 
                class="gap-x-4 w-full items-center grid" 

                style="
                    grid-template-columns: minmax(80px, 1fr) {
                        !collapseGrid ?
                        (previewView === 'example' ? "minmax(108px, 164px)" : `minmax(auto, 168px)`) :
                        "0px"
                    };
                "
            >
                {#if sortedProfile}
                    {#each sortedProfile as column (column.name)}
                    <div 
                        class="font-medium break-word grid gap-x-2 items-center" 
                        style="
                            height: 19px;
                            grid-template-columns: max-content minmax(90px, max-content);
                        " 
                        bind:this={colSizer}>
                        {#if column.summary}
                        <div
                        in:fade 
                        title={typeToSymbol(column.type).text}
                        class:bg-sky-100={categoricals.has(column.type)}
                        class:bg-red-100={!categoricals.has(column.type)}
                        class:bg-teal-200={column.conceptualType === 'TIMESTAMP' || column.type === 'TIMESTAMP'}
                        class:text-sky-800={categoricals.has(column.type)}
                        class:text-red-800={!categoricals.has(column.type)}
                        class:text-teal-800={column.conceptualType === 'TIMESTAMP' || column.type === 'TIMESTAMP'}
                        class="
                            text-ellipsis overflow-hidden whitespace-nowrap 
                            grid place-items-center rounded" 
                            style="font-size:8px; width: 16px; height: 16px;">
                            <div style="transform: translateY(.5px);">
                                {typeToSymbol(column.type).symbol}                           
                            </div> 
                        </div>
                        {:else}
                            <div in:fade class="grid place-items-center" style="width: 16px; height: 16px;">
                                <Spinner size=".45rem" bg="hsl(240, 1%, 70%)" />
                            </div>
                        {/if}
                        <button
                            disabled={!column.summary}
                            title={column.name}
                            class='text-ellipsis overflow-hidden whitespace-nowrap break-all text-left {(selectingColumns || showSummaries) ? 'hover:underline' : ''}' 
                            class:text-gray-500={!column.summary}
                            class:italic={!column.summary}
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
                            {#if column.conceptualType === 'TIMESTAMP' && column?.summary?.interval}<span class='text-gray-500 italic pl-2' style="font-size:11px">{ intervalToTimestring(column.summary.interval)}</span>{/if}
                        </button>

                    </div>
                    <!-- Preview elements -->
                    <div style:max-width="{108 + 68}px" class="justify-self-end">
                        {#if !collapseGrid}
                            <div  class="grid" style:grid-template-columns="max-content max-content">
                            {#if previewView === 'card'}
                            <div
                                style:width="108px"
                                class='overflow-hidden whitespace-nowrap justify-self-stretch text-right text-gray-500  break-all' >
                                {#if column?.summary && (categoricals.has(column.type) || categoricals.has(column.conceptualType)) && column?.summary?.cardinality}
                                    <BarAndLabel 
                                    color={!categoricals.has(column.type) ? 'hsl(1,50%, 90%' : 'hsl(240, 50%, 90%'}
                                    value={column.summary.cardinality / cardinality}>
                                        |{formatInteger(column.summary.cardinality)}|
                                    </BarAndLabel>
                                {:else if column?.summary?.histogram}
                                    <Histogram data={column.summary.histogram} width={98} height={19} color={(column.conceptualType === 'TIMESTAMP' || column.type === 'TIMESTAMP') ? "#14b8a6" : "hsl(1,50%, 80%)"} />
                                {/if}
                            </div>
                            <div
                                style:width="68px"  
                                class='self-stretch text-right text-gray-500 break-all overflow-hidden  whitespace-nowrap '>
                                {#if cardinality && column.nullCount !== undefined}
                                    <BarAndLabel
                                        title="{column.name}: {percentage(column.nullCount / cardinality)} of the values are null"
                                        bgColor={column.nullCount === 0 ? 'bg-white' : 'bg-gray-50'}
                                        color={!categoricals.has(column.type) ? 'hsl(1,50%, 90%)' : 'hsl(240, 50%, 90%)'}
                                        value={column.nullCount / cardinality || 0}>
                                                <span class:text-gray-300={column.nullCount === 0}>∅ {percentage(column.nullCount / cardinality)}</span>
                                    </BarAndLabel>
                                {/if}
                            </div>
                            {:else}
                                <div 
                                    class='text-gray-500 italic text-right text-ellipsis overflow-hidden whitespace-nowrap ' 
                                    class:hidden={collapseGrid}>
                                    {(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>')}
                                </div>
                            {/if}
                        </div>
                        {/if}
                    </div>
                    <!-- categorical summaries -->
                    {#if showSummaries && summaryColumns.includes(column.name)}
                        <div style="grid-column: 1 / -1;" class='pt-3 pb-3 pl-3 pr-3' transition:slide|local={{duration: 200 }}>

                            {#if column?.summary && categoricals.has(column.type) || categoricals.has(column.conceptualType)}
                            <TopKSummary 
                                totalRows={cardinality}
                                topK={column.summary.topK}
                                displaySize={collapseGrid ? 'sm' : 'md'}
                            />
                            {/if}


                            {#if column?.summary && !(categoricals.has(column.type) || categoricals.has(column.conceptualType)) && column.conceptualType !=='TIMESTAMP' && column.type !== 'TIMESTAMP'}
                            <div>
                                <SummaryAndHistogram
                                    width={containerWidth - 32 - 16}
                                    height={65} 
                                    data={column.summary.histogram}
                                    min={column.summary.statistics.min}
                                    qlow={column.summary.statistics.q25}
                                    median={column.summary.statistics.q50}
                                    qhigh={column.summary.statistics.q75}
                                    mean={column.summary.statistics.mean}
                                    max={column.summary.statistics.max}
                                />
                            </div>
                            {/if}
                            {#if column?.summary && (column.conceptualType === 'TIMESTAMP' || column.type === 'TIMESTAMP') && column?.summary?.interval}
                                {JSON.stringify(column.summary.interval)}
                            {/if}
                        </div>
                    {/if}

                    {/each}
                {/if}
            </div>
        </div>
    {/if}
  </div>