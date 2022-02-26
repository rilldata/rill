<script lang="ts">
import { SvelteComponent, tick } from "svelte/internal";
import { onMount, createEventDispatcher } from "svelte";
import { fade, slide, fly } from "svelte/transition";
import { tweened } from "svelte/motion";
import { cubicInOut as easing, cubicOut } from "svelte/easing";
import { format } from "d3-format";

import Menu from "$lib/components/menu/Menu.svelte";
import MenuItem from "$lib/components/menu/MenuItem.svelte";
import * as classes from "$lib/util/component-classes";
import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

import ContextButton from "$lib/components/collapsible-table-summary/ContextButton.svelte";

import NavEntry from "$lib/components/collapsible-table-summary/NavEntry.svelte";
import TopKSummary from "$lib/components/viz/TopKSummary.svelte";

import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import Spinner from "$lib/components/Spinner.svelte";

// icons
import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";

import { dropStore } from '$lib/drop-store';
import Histogram from "$lib/components/viz/SmallHistogram.svelte";
import SummaryAndHistogram from "$lib/components/viz/SummaryAndHistogram.svelte";
import {DataTypeIcon} from "$lib/components/data-types";
import { CATEGORICALS } from "$lib/duckdb-data-types"

import SummaryViewSelector from "$lib/components/collapsible-table-summary/SummaryViewSelector.svelte";
import { defaultSort, sortByNullity, sortByCardinality, sortByName } from "$lib/components/collapsible-table-summary/sort-utils"

import { onClickOutside } from "$lib/util/on-click-outside";

import { horizontalSlide } from "$lib/transitions";

import { intervalToTimestring, formatCardinality } from "$lib/util/formatters";
import Model from "../icons/Model.svelte";

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
export let show = false;

let colSizer;

const dispatch = createEventDispatcher();

const formatInteger = format(',');
const percentage = format('.1%');

let containerWidth = 0;
let contextMenu;
let contextMenuOpen = false;
let container;

onMount(() => {
    const observer = new ResizeObserver(entries => {
        containerWidth = container?.clientWidth ?? 0;
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

let sortedProfile;
function sortByOriginalOrder() {
    sortedProfile = profile;
}

let sortMethod = defaultSort;
$: if (sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
} else {
    sortedProfile = profile;
}

let previewView = 'summaries';

// FIXME: replace once we work through logical types in parquet_scan
function typeToSymbol(fieldType) {
    if (fieldType === 'BYTE_ARRAY' || fieldType === 'VARCHAR') {
        return {symbol: "C", text: 'categorical (BYTE_ARRAY)', color: 'red'};
    } else {
        return {symbol: fieldType.slice(0,1), text: fieldType, color: 'sky'};
    }
}

// these context menu coordinates will be where the element's context menu appears.
let menuX;
let menuY;
let clickOutsideListener;
$: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
}

// state for title bar hover.
let titleElementHovered = false;

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
    <NavEntry
        expanded={show}
        selected={emphasizeTitle}
        bind:hovered={titleElementHovered}
        on:select-body={() => { 
            // show = !show; 
            // pass up select body
            dispatch('select');
        }}
        on:expand={() => {
            show = !show;
            // pass up expand
            dispatch('expand');
        }}
        {icon} >
        <span class:font-bold={emphasizeTitle} class:italic={selectingColumns}>
            {#if name.split('.').length > 1}
                {name.split('.').slice(0, -1).join('.')}<span class='text-gray-500 italic pl-1'>.{name.split('.').slice(-1).join('.')}</span>
            {:else}
                {name}
            {/if}
            {#if selectingColumns}&nbsp;<span class="font-bold"> *</span>{/if}
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
                            <span class="grid grid-flow-col gap-x-2 text-gray-500 text-clip overflow-hidden whitespace-nowrap ">
                                {#if titleElementHovered || emphasizeTitle}
                                <span ><span>{cardinality !== undefined && cardinality !== NaN ? formatInteger(interimCardinality) : "no"}</span> row{#if cardinality !== 1}s{/if}</span>
                                <span class='self-center'>
                                    <ContextButton tooltipText="delete, more..." suppressTooltip={contextMenuOpen} on:click={async (event) => { 
                                        contextMenuOpen = !contextMenuOpen;
                                        menuX = event.clientX;
                                        menuY = event.clientY;

                                        if (!clickOutsideListener) {
                                            await tick();
                                            clickOutsideListener = onClickOutside(() => {
                                                contextMenuOpen = false;
                                            }, contextMenu);
                                        }
                                        
                                        
                                    }}><MoreIcon /></ContextButton>
                                </span>
                                {/if}
                                <!-- {#if !collapseGrid && sizeInBytes !== undefined}<span style="display: inline-block; text-overflow: clip; white-space: nowrap;
                                " transition:horizontalSlide|local={{duration: 250 + sizeInBytes / 5000000 * 1.3}}>, {sizeInBytes !== undefined && $sizeTween !== NaN && sizeInBytes !== NaN ? humanFileSize($sizeTween) : ''}</span>{/if} -->
                            </span>
                        {/if}
                    </div>

            </svelte:fragment>
      </NavEntry>
    </div>
    {#if contextMenuOpen}
        <!-- place this above codemirror.-->
        <div bind:this={contextMenu}>
            <FloatingElement relationship="mouse" target={{x: menuX, y:menuY}} location="right" alignment="start">
                <Menu on:escape={()=> { contextMenuOpen = false; }} on:item-select={() => { contextMenuOpen = false; }}>
                    <MenuItem on:select={() => {
                        dispatch("delete");
                    }}>
                        delete {name}
                    </MenuItem>
                </Menu>
            </FloatingElement>
        </div>
    {/if}
    {#if show}
        <div class="pl-3 pr-5 pt-1 pb-3 pl-accordion" transition:slide|local={{duration: 120 }}>
            <div style="grid-column: 1 / -1;" class='pt-2 pb-2 flex justify-between text-gray-500'>
                <select bind:value={sortMethod} class={classes.NATIVE_SELECT}>
                    <option value={sortByOriginalOrder}>sort by original order</option>
                    <option value={defaultSort}>sort by cardinality</option>
                    <option value={sortByNullity}>sort by null %</option>
                    <option value={sortByName}>sort by name</option>
                </select>
                <select bind:value={previewView} class={classes.NATIVE_SELECT}>
                    <option value="summaries">sort by summaries</option>
                    <option value="example">sort by example</option>
                </select>
            </div>

            <!-- <SummaryViewSelector bind:sortMethod bind:previewView /> -->

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
                    <!-- FIXME: make this element work with src/lib/components/data-types/DataTypeIcon.svelte -->
                    <div 
                        class="pl-3 pr-5 font-medium break-word grid gap-x-2 items-center" 
                        style="
                            height: 19px;
                            grid-template-columns: max-content minmax(90px, max-content);
                        " 
                        bind:this={colSizer}>
                        {#if column.summary}
                            <DataTypeIcon type={column.type} />
                        <!-- <div
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
                        </div> -->
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
                            {#if previewView === 'summaries'}
                            <div
                                style:width="108px"
                                class='overflow-hidden whitespace-nowrap justify-self-stretch text-right text-gray-500  break-all' >
                                {#if column?.summary && (CATEGORICALS.has(column.type) || CATEGORICALS.has(column.conceptualType)) && column?.summary?.cardinality}
                                    <BarAndLabel 
                                    color={!CATEGORICALS.has(column.type) ? 'hsl(1,50%, 90%' : 'hsl(240, 50%, 90%'}
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
                                        color={!CATEGORICALS.has(column.type) ? 'hsl(1,50%, 90%)' : 'hsl(240, 50%, 90%)'}
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

                            {#if column?.summary && CATEGORICALS.has(column.type) || CATEGORICALS.has(column.conceptualType)}
                            <TopKSummary 
                                totalRows={cardinality}
                                topK={column.summary.topK}
                                displaySize={collapseGrid ? 'sm' : 'md'}
                            />
                            {/if}


                            {#if column?.summary && !(CATEGORICALS.has(column.type) || CATEGORICALS.has(column.conceptualType)) && column.conceptualType !=='TIMESTAMP' && column.type !== 'TIMESTAMP'}
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