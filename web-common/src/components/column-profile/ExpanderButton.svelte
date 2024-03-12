<script>
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";

  export let rotated = false;
  export let suppressTooltip = false;
</script>

<Tooltip
  location="top"
  alignment="start"
  distance={8}
  activeDelay={0}
  pad={8}
  suppress={suppressTooltip}
>
  <div class="py-1 h-full">
    <button
      on:click|preventDefault|stopPropagation
      style:grid-column="left-control"
      class:-rotate-90={!rotated}
      class="
    
    h-full
    flex items-center justify-center
    flex-none
    aspect-square
    focus:outline-none
    focus:bg-gray-300
    hover:bg-gray-300
    transition-tranform
    text-gray-400
    duration-100
    rounded
    "
    >
      <CaretDownIcon size="14px" />
    </button>
  </div>
  <TooltipContent slot="tooltip-content">
    <SlidingWords active={rotated}>columns</SlidingWords>
  </TooltipContent>
</Tooltip>
