<script lang="ts">
  export let cursor: { x: number; y: number } | null;
  export let snapped: boolean;

  $: x1 = cursor?.x;
  $: y1 = cursor?.y;
</script>

{#if (x1 && y1) || x1 === 0 || y1 === 0}
  <line {x1} x2={x1} y1="0" y2="100%" class:snapped />

  <line {y1} y2={y1} x1="0" x2="100%" class:snapped />
{/if}

<style lang="postcss">
  line {
    @apply stroke-gray-300;
    stroke-color: gray;
    pointer-events: none;
    stroke-width: 1px;
    stroke-dasharray: 2.5px;
    vector-effect: non-scaling-stroke;
  }

  .snapped {
    @apply stroke-primary-400;
  }
</style>
