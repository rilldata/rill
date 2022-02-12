<script lang="ts">
/**
 * TableCell.svelte
 * notes:
 * - the max cell-width that preserves a timestamp is 210px.
*/
import { createEventDispatcher } from "svelte";
import { timeFormat } from "d3-time-format"
export let type;
export let value;
export let name;
export let index;

const dispatch = createEventDispatcher();

/**
 * FIXME: should we format the date according to the range?
 * IF date and time varies, we show with same styling
 * IF date differs but time does not, we gray out time
 * IF time differs but date does not, we gray out date.
 * For now, et's just default to showing the value.
 */
let timeFormatter = timeFormat('%b %d, %Y %I:%M:%S');

let formattedValue;
$: {
    if (type === 'TIMESTAMP') {
        formattedValue = timeFormatter(value);
    } else if(value === null) {
        formattedValue = `∅ null`
    } else {
        formattedValue = value//`val: ${value}`
    }
}

let styleType:string;
let isNull = false;
$: {
    if (value === null) {
        isNull = true;
    }
    if (type === 'VARCHAR') {
        styleType = 'varchar';
    } else if (type === 'TIMESTAMP') {
        styleType = 'timestamp';
    } else {
        styleType = 'number';
    }
}

let activeCell = false;

</script>

<td
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
        border
        border-gray-200
        {styleType}
        {isNull && 'null'}
        {activeCell && 'bg-gray-200'}
    "
    style:width="var(--table-column-width-{name}, 210px)"
    style:max-width="var(--table-column-width-{name}, 210px)"
>
    {formattedValue}
</td>

<style lang="postcss">

td {
    @apply text-gray-700;
}

.null {
    @apply text-gray-400 italic;
}

.number {
    @apply pl-8 text-right font-semibold;
}

.null.number {
    @apply text-right font-normal;
}

.timestamp {
    @apply pl-8 text-right italic font-semibold text-slate-500;
}

</style>