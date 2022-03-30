<script>
import { createEventDispatcher } from "svelte";
import { createShiftClickAction } from "$lib/util/shift-click-action"

const dispatch = createEventDispatcher();
const { shiftClickAction } = createShiftClickAction();

export let active = false;
export let emphasize = false;
export let hideRight = false;

export let left = 8 // "pl-8 pl-10";
export let right = 4 // pr-2";

</script>

<div>
<button 
    class="
        pl-{left} pr-{right}
        select-none	
        flex 
        space-between 
        gap-2
        hover:bg-gray-100 
        focus:bg-gray-100
        focus:ring-gray-500
        focus:outline-gray-300 flex-1
        justify-between w-full"
    class:bg-gray-50={active}
    use:shiftClickAction
    on:shift-click
    on:click={(event) => { dispatch('select'); }}
>
    <div class="flex gap-2 grow items-baseline flex-1" style:min-width="0px">
        <div class="self-center flex items-center">
            <slot name="icon"></slot>
        </div>
        <div class:font-bold={emphasize} class="justify-items-stretch shrink w-full text-left flex-1" style:min-width="0px">
            <slot name="left" />
        </div>
    </div>
    <div class:hidden={hideRight} class="flex gap-2 items-center">
        <slot name="right" />
        <slot name="context-button" />
    </div>
</button>
<div class="w-full">
    <slot name="details" />
</div>

</div>