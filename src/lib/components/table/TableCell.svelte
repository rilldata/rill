<script lang="ts">
/**
 * TableCell.svelte
 * notes:
 * - the max cell-width that preserves a timestamp is 210px.
*/
import { createEventDispatcher } from "svelte";
import { fade } from "svelte/transition"
import { FormattedDataType } from "$lib/components/data-types/";
import { TIMESTAMPS } from "$lib/duckdb-data-types";
import { standardTimestampFormat } from "$lib/util/formatters";
export let type;
export let value;
export let name;
export let index = undefined;
export let isNull = false;


const dispatch = createEventDispatcher();
/**
 * FIXME: should we format the date according to the range?
 * IF date and time varies, we show with same styling
 * IF date differs but time does not, we gray out time
 * IF time differs but date does not, we gray out date.
 * For now, let's just default to showing the value.
 */

let formattedValue;
$: {
    if (TIMESTAMPS.has(type)) {
        // FIXME: apparently timestamp columns get returned as strings.
        formattedValue = standardTimestampFormat(typeof value === 'string' ? new Date(value) : value, type);
    } else if(value === null) {
        formattedValue = `∅ null`
    } else {
        formattedValue = value//`val: ${value}`
    }
}


let activeCell = false;

</script>

<div
    on:mouseover={() => { dispatch('inspect', index); activeCell = true; }}
    on:mouseout={() => { activeCell = false; }}
    on:focus={() => { dispatch('inspect', index); activeCell = true; }}
    on:blur={() => {activeCell = false;} }
    title={value}
    class="
        text-ellipsis overflow-hidden whitespace-nowrap
        p-2
        pl-4
        pr-4
        border-b border-r
        border-gray-200
        {activeCell && 'bg-gray-200'}
    "
    style:width="var(--table-column-width-{name}, 210px)"
>
    {#if value !== undefined}
    <span transition:fade|local={{duration: 75}}>
        <FormattedDataType {type} {isNull} inTable>
            {formattedValue}
        </FormattedDataType>
    </span>
    {/if}
</div>