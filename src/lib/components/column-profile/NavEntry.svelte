<script lang="ts">
// FIXME: we should probably rename this to AssetNavigationElement.svelte or something like that.
import { createEventDispatcher } from "svelte";
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
import ExpanderButton from "$lib/components/column-profile/ExpanderButton.svelte";
import type { SvelteComponent } from "svelte";

const dispatch = createEventDispatcher();

export let expanded = true;
export let expandable = true;
export let selected = false;
export let hovered = false;
export let icon:SvelteComponent;
</script>

<div
    on:mouseenter={() => { hovered = true; }}
    on:mouseleave={() => { hovered = false; }}
    style:height="24px"
    style:grid-template-columns="[left-control] max-content [body] auto"
    class="
        {selected ? 'bg-gray-100' : 'bg-transparent'}
        grid
        grid-flow-col
        gap-1
        items-center
        hover:bg-gray-200
        pl-3 pr-5 
    "
>
    {#if expandable}
    <ExpanderButton rotated={expanded} on:click={() => { dispatch('expand' )}}>
        <CaretDownIcon size="14px" />
    </ExpanderButton>
    {/if}
    <button 
        on:click={() => { dispatch('select-body')}}
        on:focus={() => { hovered = true; }}
        on:blur={() => { hovered = false; }}
        style:grid-column="body"
        style:grid-template-columns="[icon] max-content [text] 1fr [contextual-information] max-content"
        class="
            w-full 
            justify-start
            text-left 
            grid  gap-2
            items-center
            p-0"
        >
        <!-- {#if icon}
            <div 
                style:grid-column="icon" 
                class="text-gray-400">
                <svelte:component this={icon} size=14px />
            </div>
        {/if} -->
        <div
            style:grid-column="text"
            class="
                justify-self-auto
                text-ellipsis 
                overflow-hidden 
                whitespace-nowrap">
            <slot />
        </div>

        <div 
            style:grid-column="contextual-information"
            class="justify-self-end"
        >
            <slot name="contextual-information" />
        </div>
    </button>
</div>