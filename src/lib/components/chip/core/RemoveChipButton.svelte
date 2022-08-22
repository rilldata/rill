<script lang="ts">
  import CancelCircle from "$lib/components/icons/CancelCircle.svelte";

  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { createEventDispatcher, getContext } from "svelte";
  import type { Writable } from "svelte/store";

  export let tooltipLocation = "bottom";
  export let tooltipAlignment = "start";
  export let tooltipDistance = 12;

  const tooltipSuppression = getContext(
    "rill:app:childRequestedTooltipSuppression"
  ) as Writable<boolean>;
  const dispatch = createEventDispatcher();

  function focusOnRemove() {
    tooltipSuppression.set(true);
  }
  function blurOnRemove() {
    tooltipSuppression.set(false);
  }
</script>

<Tooltip
  location={tooltipLocation}
  alignment={tooltipAlignment}
  distance={tooltipDistance}
>
  <button
    on:mouseover={focusOnRemove}
    on:focus={focusOnRemove}
    on:mouseleave={blurOnRemove}
    on:blur={blurOnRemove}
    on:click|stopPropagation={() => dispatch("remove")}
  >
    <CancelCircle size="16px" />
    <!-- <Close /> -->
  </button>
  <div slot="tooltip-content">
    {#if $$slots["remove-tooltip"]}
      <TooltipContent slot="tooltip-content" maxWidth="300px">
        <slot name="remove-tooltip" />
      </TooltipContent>
    {/if}
  </div>
</Tooltip>
