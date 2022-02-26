<script lang="ts">
/** This component is newer than the CollapsibleTitle, which should primarily be used
 * for expandable asset elements. By contrast, this element is used as a header element
 * for asset sections.
*/
import { onMount } from "svelte";
import { fly } from "svelte/transition";
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";

import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

import type { SvelteComponent } from "svelte";
export let active = true;
export let icon:SvelteComponent;
let duration = 0;

let tooltipActive;

// here's where we handle setting the active tooltip
// duration, which will help us animate the fly element.
let timer;
$: if (tooltipActive) {
    timer = setTimeout(() => {
        duration = 150;
    }, 200);
    
} else {
    clearTimeout(timer);
    duration = 0;
}
</script>

<div 
    class='collapsible-title align grid grid-cols-max justify-between'
    style="grid-template-columns: auto max-content;"
>
<Tooltip location="right" alignment="middle" distance={8} bind:active={tooltipActive}>
    <button 
        class="
            bg-transparent 
            grid 
            grid-flow-col 
            gap-2
            items-center
            p-0 
            pr-1 
            border-transparent 
            hover:border-slate-200"
            style="
                max-width: 100%;

            "
        on:click={() => { active = !active; }}>
        <!-- {#if icon}
            <div class="text-gray-400">
                <svelte:component this={icon} size=1em />
            </div>
        {/if} -->
        <div class="text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 font-bold uppercase">
            <slot />
        </div>
    </button>
    <TooltipContent slot="tooltip-content">
        <div class="relative">
            <span class="invisible">{active ? "hide" : "show"}</span>
            {#key active}
                <div class="absolute" style:left="0" style:top="0px" transition:fly={{duration, y: 7.5 * (!active ? 1 : -1)}}>{active ? "hide" : "show"}</div>
            {/key}
        </div>
        
    </TooltipContent>
</Tooltip>
    <div class="contextual-information justify-self-stretch text-right">
        <slot name="contextual-information" />
    </div>
</div>