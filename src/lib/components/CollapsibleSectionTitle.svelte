<script lang="ts">
/** This component is newer than the CollapsibleTitle, which should primarily be used
 * for expandable asset elements. By contrast, this element is used as a header element
 * for asset sections.
*/

import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
import type { SvelteComponent } from "svelte";
export let active = true;
export let icon:SvelteComponent;
</script>

<div 
    class='collapsible-title align grid grid-cols-max justify-between'
    style="grid-template-columns: auto max-content;"
>
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

    <div class="contextual-information justify-self-stretch text-right">
        <slot name="contextual-information" />
    </div>
</div>