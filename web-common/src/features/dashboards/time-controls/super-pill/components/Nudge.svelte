<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let direction: "left" | "right";
  export let onPan: (direction: "left" | "right") => void;
</script>

<button
  disabled={direction === "left" ? !canPanLeft : !canPanRight}
  on:click={() => onPan(direction)}
>
  <Tooltip distance={8} location="top">
    <span class={direction}>
      <CaretDownIcon size="16px" />
    </span>
    <TooltipContent slot="tooltip-content">
      Step time {direction === "left" ? "back" : "forward"}
    </TooltipContent>
  </Tooltip>
</button>

<style lang="postcss">
  button {
    @apply w-8 flex items-center justify-center relative;
    @apply transform;
  }

  .left {
    @apply rotate-90;
  }

  .right {
    @apply -rotate-90;
  }

  button:disabled {
    @apply cursor-not-allowed;
    @apply opacity-50;
  }

  span {
    @apply h-full flex items-center;
  }
</style>
