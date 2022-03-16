<script lang="ts">
/**
 * PreviewTable.svelte
 * Use this component to drop into the application.
 * Its goal it so utilize all of the other container components
 * and provide the interactions needed to do things with the table.
*/
import { slide } from "svelte/transition";
import { Table, TableRow, TableHeader, TableCell } from "$lib/components/table/";
import { FormattedDataType, DataTypeIcon } from "$lib/components/data-types/";
import Pin from "$lib/components/icons/Pin.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import DataTypeTitle from "$lib/components/tooltip/DataTypeTitle.svelte";
import PreviewTableHeader from "./PreviewTableHeader.svelte";
import { TIMESTAMPS } from "$lib/duckdb-data-types";
import { standardTimestampFormat } from "$lib/util/formatters"

interface ColumnName {
    name:string;
    type:string;
}

export let columnNames:ColumnName[];
export let rows:any[];

let visualCellField = undefined;
let visualCellValue = undefined;
let visualCellType = undefined;

let selectedColumns = [];

let activeIndex;

function setActiveElement(value, name, index) {
    visualCellValue = value; 
    visualCellField = name;
    visualCellType = columnNames.find(column => column.name === visualCellField).type
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

<div class='flex relative bg-gray-100'>
    
    <Table on:mouseleave={() => { visualCellValue = undefined; }}>
        <!-- headers -->
        <TableRow>
            {#each columnNames as {name, type} (name)}
                {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
                <PreviewTableHeader 
                    {name} 
                    {type} 
                    pinned={thisColumnIsPinned} 
                    on:pin={() => {
                        togglePin(name, type, selectedColumns);
                    }}
                />
            {/each}
        </TableRow>
        <!-- values -->
        {#each rows as row, index}
            <TableRow hovered={activeIndex === index}>
                {#each columnNames as { name, type } (index+name)}
                    <TableCell
                        on:inspect={() => { setActiveElement(row[name], name, index) }} 
                        {name} 
                        {type} 
                        value={row[name]}
                        isNull={row[name] === null}

                    />
                {/each}
            </TableRow>
        {/each}
    </Table>

    {#if selectedColumns.length}
    <div class="sticky right-0 z-20 bg-white border border-l-4 border-t-0 border-b-0 border-r-0 border-gray-300" >
        <Table>
            <TableRow>
                {#each selectedColumns as {name, type} (name)}
                    {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
                    <PreviewTableHeader 
                        {name} 
                        {type} 
                        pinned={thisColumnIsPinned} 
                        on:pin={() => {
                            togglePin(name, type, selectedColumns);
                        }}
                    />
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
                            isNull={row[name] === null}
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
        class="sticky bottom-0 left-0 bg-white p-3 border border-t-1 border-gray-200 pointer-events-none z-30 grid grid-flow-col justify-start gap-x-3 items-baseline"
        style:box-shadow="0 -4px 2px 0 rgb(0 0 0 / 0.05)"
    >
        <span class='font-bold pr-5'>{visualCellField}</span>
        <FormattedDataType type={visualCellType} isNull={visualCellValue === null}>
            {TIMESTAMPS.has(visualCellType) ? standardTimestampFormat(new Date(visualCellValue), visualCellType) : visualCellValue}
        </FormattedDataType>
</div>
{/if}
