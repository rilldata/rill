<script lang="ts">
  export let width: number | undefined;
  export let height: number | undefined;
  export let x: number | undefined;
  export let y: number | undefined;
</script>

<div
  class="grid-stack-item"
  data-gs-w={width}
  data-gs-h={height}
  data-gs-x={x}
  data-gs-y={y}
>
  <div class="grid-stack-item-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .grid-stack-item-content {
    @apply flex flex-col items-center justify-center;
    @apply bg-white border border-gray-200 rounded-md shadow-sm;
  }
</style>
