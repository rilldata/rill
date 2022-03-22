<script lang="ts">
import { SvelteComponent, tick } from "svelte/internal";
import { onMount, createEventDispatcher } from "svelte";
import { slide } from "svelte/transition";
import { tweened } from "svelte/motion";
import { cubicInOut as easing, cubicOut } from "svelte/easing";
import { format } from "d3-format";

import Menu from "$lib/components/menu/Menu.svelte";
import MenuItem from "$lib/components/menu/MenuItem.svelte";
import * as classes from "$lib/util/component-classes";
import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

import ColumnProfile from "./ColumnProfile.svelte";

import Spacer from "$lib/components/icons/Spacer.svelte";

import NavEntry from "$lib/components/column-profile/NavEntry.svelte";

// icons
import MoreIcon from "$lib/components/icons/MoreHorizontal.svelte";

import { dropStore } from '$lib/drop-store';

import { defaultSort, sortByNullity, sortByCardinality, sortByName } from "$lib/components/column-profile/sort-utils"

import { onClickOutside } from "$lib/util/on-click-outside";
import { config } from "./utils";

export let icon:SvelteComponent;
export let name:string;
export let path:string;
export let cardinality:number;
export let profile:any;
export let head:any; // FIXME
export let sizeInBytes:number;
export let emphasizeTitle:boolean = false;
export let draggable = true;
export let show = false;
export let showTitle = true;
export let showContextButton = true;
export let indentLevel = 0;

const dispatch = createEventDispatcher();

const formatInteger = format(',');

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


let cardinalityTween = tweened(cardinality, { duration: 600, easing });
let sizeTween = tweened(sizeInBytes, { duration: 650, easing, delay: 150 });

$: cardinalityTween.set(cardinality || 0);
$: interimCardinality = ~~$cardinalityTween;
$: sizeTween.set(sizeInBytes || 0);

let selectingColumns = false;
let selectedColumns = [];

let sortedProfile;
function sortByOriginalOrder() {
    sortedProfile = profile;
}

let sortMethod = defaultSort;
// this predicate actually is valid but typescript doesn't seem to agree.
// @ts-ignore
$: if (sortMethod !== sortByOriginalOrder) {
    sortedProfile = [...profile].sort(sortMethod);
} else {
    sortedProfile = profile;
}

let previewView = 'summaries';

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
    {#if showTitle}
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
                                    <ContextButton tooltipText="delete" suppressTooltip={contextMenuOpen} on:click={async (event) => { 
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
    {/if}

    {#if show}
        <div class="pt-1 pb-3 pl-accordion" transition:slide|local={{duration: 120 }}>
            <!-- pl-16 -->
            <div  class='pl-{indentLevel === 1 ? '10' : '4'} pr-5 pt-2 pb-2 flex justify-between text-gray-500' class:flex-col={containerWidth < 325}>
                <select style:transform="translateX(-4px)" bind:value={sortMethod} class={classes.NATIVE_SELECT}>
                    <option value={sortByOriginalOrder}>show original order</option>
                    <option value={defaultSort}>sort by type</option>
                    <option value={sortByNullity}>sort by null %</option>
                    <option value={sortByName}>sort by name</option>
                </select>
                <select style:transform="translateX(4px)" bind:value={previewView} class={classes.NATIVE_SELECT} class:hidden={containerWidth < 325}>
                    <option value="summaries">show summary&nbsp;</option>
                    <option value="example">show example</option>
                </select>
            </div>

            <div>
                {#if sortedProfile && sortedProfile.length && head.length}
                    {#each sortedProfile as column (column.name)}
                    <ColumnProfile
                        {indentLevel}
                        example={head[0][column.name] || ''}
                        containerWidth={containerWidth}

                        hideNullPercentage={containerWidth < config.hideNullPercentage}
                        hideRight={containerWidth < config.hideRight}
                        compactBreakpoint={config.compactBreakpoint}
                        view={previewView}
                        name={column.name}
                        type={column.type}
                        summary={column.summary}
                        totalRows={cardinality}
                        nullCount={column.nullCount}
                    >
                        <button slot="context-button" class:hidden={!showContextButton}>
                                <Spacer size="16px" />
                        </button>
                    </ColumnProfile>

                    
                    {/each}
                {/if}
            </div>
        </div>
    {/if}
  </div>