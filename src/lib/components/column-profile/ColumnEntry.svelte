<script>
import { createEventDispatcher } from "svelte";

const dispatch = createEventDispatcher();

export let active = false;
export let emphasize = false;
export let hideRight = false;

export let left = 8 // "pl-8";
export let right = 2 // pr-2";

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
        focus:outline-gray-300
        justify-between w-full"
    class:bg-gray-50={active}
    on:click={() => {
        dispatch('select');
    }}
>
    <div class="flex gap-2 grow overflow-hidden items-baseline">
        <div class="self-center flex items-center">
            <slot name="icon"></slot>
        </div>
        <div class:font-bold={emphasize} class="text-ellipsis overflow-hidden whitespace-nowrap">
            <slot name="left" />
        </div>
    </div>
    <div class:hidden={hideRight} class="flex gap-2 items-center">
        <slot name="right" />
        <slot name="context-button" />
    </div>
</button>
<div>
    <slot name="details" />
</div>

</div>