<script lang="ts">
  import CancelCircle from "../../icons/CancelCircle.svelte";

  import { createEventDispatcher, getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import { defaultChipColors } from "../chip-types";

  export let tooltipLocation = "bottom";
  export let tooltipAlignment = "start";
  export let tooltipDistance = 12;
  export let textClass = defaultChipColors.textClass;
  export let supressTooltip = false;

  const tooltipSuppression = getContext(
    "rill:app:childRequestedTooltipSuppression",
  ) as Writable<boolean>;
  const dispatch = createEventDispatcher();

  function focusOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(true);
  }
  function blurOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(false);
  }
</script>

<Tooltip
  location={tooltipLocation}
  alignment={tooltipAlignment}
  distance={tooltipDistance}
  suppress={supressTooltip}
>
  <button
    class="{textClass} pl-2"
    on:mouseover={focusOnRemove}
    on:focus={focusOnRemove}
    on:mouseleave={blurOnRemove}
    on:blur={blurOnRemove}
    on:click|stopPropagation={() => dispatch("remove")}
    aria-label="Remove"
  >
    <CancelCircle size="16px" />
    <!-- <Close /> -->
  </button>
  <div slot="tooltip-content">
    {#if $$slots["remove-tooltip"]}
      <TooltipContent maxWidth="300px">
        <slot name="remove-tooltip" />
      </TooltipContent>
    {/if}
  </div>
</Tooltip>
