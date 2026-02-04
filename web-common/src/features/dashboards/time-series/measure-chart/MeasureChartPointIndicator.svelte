<script lang="ts">
  export let x: number;
  export let y: number;
  export let zeroY: number;
  export let value: string;
  export let selected: boolean = false;
</script>

<line
  class="stroke-fg-muted"
  class:stroke-theme-500={selected}
  stroke-width={selected ? 2 : 1.5}
  x1={x}
  x2={x}
  y1={zeroY}
  y2={y}
/>

<circle
  class="stroke-[3px] fill-theme-600"
  class:fill-theme-500={selected}
  cx={x}
  cy={y}
  r={4}
/>

<text
  class="text-xs font-semibold fill-fg-secondary"
  class:fill-theme-500={selected}
  style:paint-order="stroke"
  stroke-width="3px"
  x={x + 8}
  y={y + 4}
>
  {value}
</text>

<style lang="postcss">
  circle,
  text {
    @apply stroke-surface-background;
    paint-order: stroke;
  }
</style>
