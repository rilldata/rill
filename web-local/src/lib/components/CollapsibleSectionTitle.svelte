<script lang="ts">
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

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
        class="text-ellipsis overflow-hidden whitespace-nowrap text-gray-500 font-medium"
      >
        <slot />
      </div>
    </button>
    <TooltipContent slot="tooltip-content">
      <SlidingWords {active}>
        {tooltipText}
      </SlidingWords>
    </TooltipContent>
  </Tooltip>
  <div class="contextual-information justify-self-stretch text-right">
    <slot name="contextual-information" />
  </div>
</div>
