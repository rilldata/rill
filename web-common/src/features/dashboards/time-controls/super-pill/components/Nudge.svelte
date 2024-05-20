<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let onPan: (direction: "left" | "right") => void;
  export let direction: "left" | "right";
</script>

<button
  disabled={direction === "left" ? !canPanLeft : !canPanRight}
  title="Pan time range {direction === 'left' ? 'back' : 'forward'}"
  on:click={() => onPan(direction)}
>
  <span class={direction}>
    <CaretDownIcon size="16px" />
  </span>
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
</style>
