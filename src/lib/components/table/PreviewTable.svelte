<script lang="ts">
/**
 * PreviewTable.svelte
 * Use this component to drop into the application.
 * Its goal it so utilize all of the other container components
 * and provide the interactions needed to do things with the table.
*/
import { slide } from "svelte/transition";
import { Table, TableRow, TableHeader, TableCell } from "$lib/components/table/";
import { DataType } from "$lib/components/data-types/"
import Pin from "$lib/components/icons/Pin.svelte";

interface ColumnName {
    name:string;
    type:string;
}

export let columnNames:ColumnName[];
export let rows:any[];

let visualCellField = undefined;
let visualCellValue = undefined;

let selectedColumns = [];

let activeIndex;

function setActiveElement(value, name, index) {
    visualCellValue = value; 
    visualCellField = name;
    activeIndex = index;
}

function columnIsPinned(name, selectedCols) {
    return selectedCols.map(column => column.name).includes(name);
}

function togglePin(name, type, selectedCols) {
    // if column is already pinned, remove.
    if (columnIsPinned(name, selectedCols)) {
        selectedColumns = [...selectedCols.filter(column => column.name !== name)]
    } else {
        selectedColumns = [...selectedCols, {name, type}];
    }
}

</script>

<div class='flex relative'>
    
    <Table on:mouseleave={() => { visualCellValue = undefined; }}>
        <!-- headers -->
        <TableRow header>
            {#each columnNames as {name, type}}
                {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
                <TableHeader {name} {type}>
                 <div class="
                    grid
                    grid-flow-col
                    justify-between
                    gap-x-3">

                    <div>
                        {name}
                    </div>
                    <button 
                        class:text-gray-900={thisColumnIsPinned} 
                        class:text-gray-400={!thisColumnIsPinned}
                        class="transition-colors duration-100"
                        on:click={() => { togglePin(name, type, selectedColumns)}}
                        >
                        <Pin size=".9em" />
                    </button>
                </div>
                 </TableHeader>
            {/each}
        </TableRow>
        <!-- values -->
        {#each rows as row, index}
            <TableRow hovered={activeIndex === index}>
                {#each columnNames as { name, type }}
                    <TableCell
                    on:inspect={() => { setActiveElement(row[name], name, index) }} 
                    {name} 
                    {type} 
                    value={row[name]}
                    />
                {/each}
            </TableRow>
        {/each}
    </Table>

    {#if selectedColumns.length}
    <div class="sticky right-0 z-20 bg-white border border-l-4 border-t-0 border-b-0 border-r-0 border-gray-400" >
        <Table>
            <TableRow header>
                {#each selectedColumns as {name, type} (name)}
                    <TableHeader {type}>
                        <div class="
                        grid
                        grid-flow-col
                        justify-between
                        gap-x-3">
    
                        <div>
                            {name}
                        </div>
                        <button class="text-gray-900" on:click={() => { togglePin(name, type, selectedColumns)}}>
                            <Pin size=".9em" />
                        </button>
                    </div>
                    </TableHeader>
                {/each}
            </TableRow>
                {#each rows as row, index}
                <TableRow hovered={activeIndex === index}>
                    {#each selectedColumns as { name, type }}
                        <TableCell 
                            on:inspect={() => { setActiveElement(row[name], name, index) }}
                            {name} 
                            {type} 
                            {index}
                            value={row[name]} />
                    {/each}
                </TableRow>
            {/each}
        </Table>
    </div>
    {/if}
</div>
{#if visualCellValue !== undefined}
<div 
    transition:slide={{duration: 100}} 
        class="sticky bottom-0 left-0 bg-white p-3 border border-t-1 border-gray-200 pointer-events-none z-30"
        style:box-shadow="0 -4px 2px 0 rgb(0 0 0 / 0.05)"
    >
        <span class='font-bold pr-5'>{visualCellField}</span>
        <DataType type={columnNames.find(column => column.name === visualCellField).type} isNull={visualCellValue === null}>
            {visualCellValue}
        </DataType>
</div>
{/if}
