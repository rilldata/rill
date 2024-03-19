<script lang="ts">
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";

  export let excludeMode: boolean;
  export let isBeingCompared: boolean;
  export let sorted: boolean;
  export let selected: boolean;

  export let value: string | number | boolean | undefined;
  export let formattedValue: string | null = null;
  export let type: string;
  export let total: number | null = null;
</script>

<div
  class:sorted
  class="size-full px-2 py-1 flex items-center relative"
  class:selected
  class:no-data={!value}
  class:ui-copy-number={type === "INT"}
>
  <p class=" w-full text-right">
    {!value ? "no data" : formattedValue || value}
  </p>

  {#if total}
    <span
      class:selected
      class="bar"
      style:--bar-size="{(Number(value) / total) * 100}%"
    >
    </span>
  {/if}
</div>

<style lang="postcss">
  div {
    --bar-color: var(--color-primary-100);
  }
  .selected {
    @apply font-bold;
    --bar-color: var(--color-primary-200);
  }

  .sorted {
    @apply bg-gray-50;
  }

  .no-data {
    @apply text-gray-400 italic;
    font-size: 0.925em;
  }

  .bar {
    @apply absolute top-0 left-0 z-0 size-full mix-blend-multiply;
    background: linear-gradient(
      to right,
      var(--bar-color) var(--bar-size),
      transparent 0%
    );
  }
</style>
