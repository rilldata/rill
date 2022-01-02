<script lang="ts">
import { onMount } from "svelte";
import { slide } from "svelte/transition";
import { tweened } from "svelte/motion";
import { cubicInOut as easing } from "svelte/easing";
import { format } from "d3-format";

import CollapsibleTitle from "$lib/components/CollapsibleTitle.svelte";

import { dropStore } from '$lib/drop-store';

export let name:string;
export let path:string;
export let cardinality:number;
export let profile:any;
export let head:any;
export let sizeInBytes:number;
export let emphasizeTitle:boolean = false;

let colSizer;

const formatCardinalityFcn = format(".3s");
function formatCardinality(n) {
    let fmt:Function;
    if (n <= 1000) {
        fmt = format(',');
        return fmt(~~n);
    } else {
        fmt = format('.3s');
        return fmt(n);
    }
}
let container;

let containerWidth = 0;
let firstColWidth = 0;
let show = false;

function humanFileSize(size:number) {
    var i = Math.floor( Math.log(size) / Math.log(1024) );
    return ( size / Math.pow(1024, i) ).toFixed(2) + ['B', 'K', 'M', 'G'][i];
};

onMount(() => {
    const observer = new ResizeObserver(entries => {
        entries.forEach((entry) => {
            containerWidth = entry.target.clientWidth;
        });
        firstColWidth = colSizer.offsetWidth;
    });
    observer.observe(container);
});

//let collapseGrid = false;
//$: collapseGrid = containerWidth - 120 - 16 / firstColWidth < 1;
$: collapseGrid = containerWidth < 200 + 120 + 16;

let cardinalityTween = tweened(cardinality, { duration: 1500, easing });
let sizeTween = tweened(sizeInBytes, { duration: 1600, easing, delay: 150 });
$: cardinalityTween.set(cardinality);
$: sizeTween.set(sizeInBytes);

let draggingEditor;

let selectingColumns = false;
let selectedColumns = [];
</script>

<div bind:this={container}>
    <div draggable={true} 
        class="drag-interface"
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
            // set the drag store to be consumed elsewhere.
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
    <CollapsibleTitle bind:active={show}>
        <span class:font-bold={emphasizeTitle} class:italic={selectingColumns}>
            {name}{#if selectingColumns}&nbsp;<span class="font-bold"> *</span>{/if}
        </span>
            <svelte:fragment slot='contextual-information'>
                    <div class='italic'>
                        {#if selectingColumns}
                            select columns
                        {:else}
                            {formatCardinality($cardinalityTween)} row{#if cardinality !== 1}s{/if}{#if !collapseGrid}, {humanFileSize($sizeTween)}{/if}
                        {/if}
                        <button class:font-bold={selectingColumns} on:click={() => {
                            selectingColumns = !selectingColumns;
                        }}>*</button>
                    </div>

            </svelte:fragment>
      </CollapsibleTitle>
    </div>
    {#if show}
        <div class="pt-1 pl-accordion" transition:slide|local={{duration: 120 }}>
            <!-- {#if path}
                <div class='pb-2 pt-2 italic'>{path}</div>
            {/if} -->
            <!-- {#if collapseGrid}
            <div class=" pb-1 italic">
                {formatCardinality(cardinality)} row{#if cardinality !== 1}s{/if}
                {humanFileSize(sizeInBytes)}
            </div>
            {/if} -->
            <!-- {#if selectingColumns}
                {selectedColumns.join(',')}
            {/if} -->
            <div class="rows" class:break-grid={collapseGrid}>
                {#each profile as column}
                <div class="font-medium break-word">
                    <button class='break-all {selectingColumns ? 'hover:underline' : ''}' class:font-bold={selectingColumns && selectedColumns.includes(column.name)} 
                        on:click={() => {
                        if (selectingColumns) {
                            if (selectedColumns.includes(column.name)) {
                                console.log('removing', column.name)
                                selectedColumns = selectedColumns.filter(c => c !== column.name)
                            } else {

                                console.log('adding', column.name)
                                selectedColumns = [...selectedColumns, column.name];
                            }
                        }
                    }}>
                        {column.name} 
                    </button>
                    <span 
                        class="text-gray-500"
                    >
                        {column.type}
                    </span>
                    <span class="font-light text-gray-500">
                    {#if column.pk === 1} (primary){:else}{/if}
                </span>
                </div>
                <div class='justify-self-end text-right text-gray-500 italic break-all' class:remove={collapseGrid}>
                    {(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>').slice(0,25)}
                </div>
                {/each}
            </div>
        </div>
    {/if}

    <!-- <table cellpadding="0" cellspacing="0">
    {#each profile as column}
      <tr>
        <td>
        <div class="font-medium">{column.name} 
            <span class="column-type">
              {column.type}
            </span>
            <span class="font-light text-gray-500">
            {#if column.pk === 1} (primary){:else}{/if}
          </span></div> 
        </td>
        <td class='column-example'>
          {(head[0][column.name] !== '' ? `${head[0][column.name]}` : '<empty>').slice(0,50)}
        </td>
      </tr>
    {/each}
    </table> -->
  </div>

<style lang="postcss">
.rows {
    display: grid;
    grid-template-columns: auto minmax(120px, auto);
    column-gap: 1rem;
    width: 100%;
}

.break-grid {
    display: block;
}

.remove {
    display: none;
}

.drag-interface:active {
    cursor: grabbing;
}

</style> 