<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../config";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";

  export let show = true;
</script>

<Tooltip location="right" alignment="middle">
  <button
    on:click={() => {
      show = !show;
    }}
    class="align-middle text-left pl-4 pr-2 font-semibold text-[11px] w-fit"
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    <SlidingWords active={show}>
      <slot />
    </SlidingWords>
  </TooltipContent>
</Tooltip>
