<script lang="ts">
// FIXME: we should probably rename this to AssetNavigationElement.svelte or something like that.
import { createEventDispatcher } from "svelte";
import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
import ExpanderButton from "$lib/components/column-profile/ExpanderButton.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
import transientBooleanStore from "$lib/util/transient-boolean-store";
import StackingWord from "$lib/components/tooltip/StackingWord.svelte";

const dispatch = createEventDispatcher();

export let expanded = true;
export let expandable = true;
export let selected = false;
export let hovered = false;

let clicked = transientBooleanStore();

</script>

<div
    on:mouseenter={() => { hovered = true; }}
    on:mouseleave={() => { hovered = false; }}
    style:height="24px"
    style:grid-template-columns="[left-control] max-content [body] auto [contextual-information] max-content"
    class="
        {selected ? 'bg-gray-100' : 'bg-transparent'}
        grid
        grid-flow-col
        gap-2
        items-center
        hover:bg-gray-200
        pl-4 pr-4 
    "
>
    {#if expandable}
    <ExpanderButton rotated={expanded} on:click={() => { dispatch('expand' )}}>
        <CaretDownIcon size="14px" />
    </ExpanderButton>
    {/if}
    <Tooltip location="right">
        <button 
            on:click={(evt) => { 
                dispatch('select-body', evt.shiftKey);
                clicked.flip();
         }}
            on:focus={() => { hovered = true; }}
            on:blur={() => { hovered = false; }}
            style:grid-column="body"
            style:grid-template-columns="[icon] max-content [text] 1fr"
            class="
                w-full 
                justify-start
                text-left 
                grid 
                items-center
                p-0"
            >
            <div
                style:grid-column="text"
                class="
                    w-full
                    justify-self-auto
                    text-ellipsis 
                    overflow-hidden 
                    whitespace-nowrap">
                <slot />
            </div>
        </button>
    <TooltipContent slot='tooltip-content'>

        <slot name="tooltip-content" />
    </TooltipContent>
    </Tooltip>
    <div 
    style:grid-column="contextual-information"
    class="justify-self-end"
>
    <slot name="contextual-information" />
</div>
</div>