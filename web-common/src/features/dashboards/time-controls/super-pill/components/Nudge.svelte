<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let onPan: (direction: "left" | "right") => void;
</script>

<button
  disabled={!canPanLeft}
  title="Pan time range backward"
  on:click={() => onPan("left")}
>
  <span class="rotate-90">
    <CaretDownIcon size="16px" />
  </span>
</button>
<button
  disabled={!canPanRight}
  title="Pan time range forward"
  on:click={() => onPan("right")}
>
  <span class="-rotate-90">
    <CaretDownIcon size="16px" />
  </span>
</button>

<style lang="postcss">
  button {
    @apply w-8 flex items-center justify-center relative;
  }

  button:disabled {
    @apply opacity-50 cursor-not-allowed;
  }
</style>
