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
import Tooltip from "../tooltip/Tooltip.svelte";
import TooltipContent from "../tooltip/TooltipContent.svelte";
import notificationStore from "../notifications";
import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
import StackingWord from "../tooltip/StackingWord.svelte";
import Shortcut from "../tooltip/Shortcut.svelte";
import TooltipTitle from "../tooltip/TooltipTitle.svelte";

import { createShiftClickAction } from "$lib/util/shift-click-action";

const { shiftClickAction } = createShiftClickAction();

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
        formattedValue = standardTimestampFormat(value, type);
    } else if(value === null) {
        formattedValue = `∅ null`
    } else {
        if (typeof value === 'string' && !value.length) {
            // replace with a whitespace chracter to preserve the cell height when we have an empty string
            formattedValue = '&nbsp;'
        } else {
            formattedValue = value;
        }
    }
}

let activeCell = false;

</script>
<Tooltip location='top' distance={16}>
<td
    on:mouseover={() => { dispatch('inspect', index); activeCell = true; }}
    on:mouseout={() => { activeCell = false; }}
    on:focus={() => { dispatch('inspect', index); activeCell = true; }}
    on:blur={() => {activeCell = false;} }
    title={value}
    class="
        p-2
        pl-4
        pr-4
        border
        border-gray-200
        {activeCell && 'bg-gray-200'}
    "
    style:width="var(--table-column-width-{name}, 210px)"
    style:max-width="var(--table-column-width-{name}, 210px)"
>
    <button 
        class="text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
        use:shiftClickAction
        on:shift-click={async (event) => {
            await navigator.clipboard.writeText(name);
            let rep = name.toString().slice(0,10)
            notificationStore.send({ message: `copied value to clipboard`});
            // update this to set the active animation in the tooltip text
        }}
    >
    {#if value !== undefined}
    <span transition:fade|local={{duration: 75}}>
        <FormattedDataType {type} {isNull} inTable>
            {@html formattedValue}
        </FormattedDataType>
    </span>
    {/if}
    </button>
</td>
<TooltipContent slot='tooltip-content'>
    <TooltipTitle>
        <svelte:fragment slot='name'>
            {value}
        </svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
        <div>
            <StackingWord>copy</StackingWord> this value to clipboard
        </div>
        <Shortcut>
            <span style='font-family: var(--system);";
            '>⇧</span> + Click
        </Shortcut>
    </TooltipShortcutContainer>
</TooltipContent>
</Tooltip>