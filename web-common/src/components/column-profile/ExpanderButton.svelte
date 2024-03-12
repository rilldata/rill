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
  pad={8}
  suppress={suppressTooltip}
>
  <span class="py-1 h-full">
    <button on:click|preventDefault|stopPropagation class:-rotate-90={!rotated}>
      <CaretDownIcon size="14px" />
    </button>
  </span>
  <TooltipContent slot="tooltip-content">
    <SlidingWords active={rotated}>columns</SlidingWords>
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  button {
    @apply h-full aspect-square rounded;
    @apply grid place-content-center;
    @apply text-gray-400;
    @apply transition-transform duration-100;
  }

  button:hover,
  button:focus {
    @apply outline-none bg-gray-300;
  }
</style>
