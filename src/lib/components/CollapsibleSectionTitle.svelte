<script lang="ts">
  /** This component is newer than the CollapsibleTitle, which should primarily be used
   * for expandable asset elements. By contrast, this element is used as a header element
   * for asset sections.
   */

  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  import SlidingWords from "$lib/components/tooltip/SlidingWords.svelte";

  export let active = true;
  export let tooltipText = "";

  let tooltipActive;
</script>

<div
  class="collapsible-title align grid grid-cols-max justify-between"
  style="grid-template-columns: auto max-content;"
>
  <Tooltip
    location="right"
    alignment="middle"
    distance={8}
    bind:active={tooltipActive}
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
      on:click={() => {
        active = !active;
      }}
    >
      <div
        class="text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 font-bold uppercase"
      >
        <slot />
      </div>
    </button>
    <TooltipContent slot="tooltip-content">
      <SlidingWords {active}>
        {tooltipText}
      </SlidingWords>
      <!-- <div class="relative">
            <span class="invisible">{active ? "hide" : "show"}</span>
                {#key active}
                    <span class="absolute" style:left="0" style:top="0px" transition:fly={{duration, y: 7.5 * (!active ? 1 : -1)}}>{active ? "hide" : "show"}</span>
                {/key}
        </div> -->
    </TooltipContent>
  </Tooltip>
  <div class="contextual-information justify-self-stretch text-right">
    <slot name="contextual-information" />
  </div>
</div>
