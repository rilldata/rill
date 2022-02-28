<script>
import { createEventDispatcher } from "svelte";

const dispatch = createEventDispatcher();

export let active = false;
export let emphasize = false;
export let hideRight = false;

</script>

<div>
<button 
    class="
        pl-8 pr-2
        select-none	flex space-between 
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
    <div class:hidden={hideRight} class="flex gap-2">
        <slot name="right" />
        <div>
            <slot name="context-button" />
        </div>
    </div>
</button>
<div>
    <slot name="details" />
</div>

</div>