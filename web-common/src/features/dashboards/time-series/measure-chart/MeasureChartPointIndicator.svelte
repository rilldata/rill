<script lang="ts">
  export let x: number;
  export let y: number;
  export let zeroY: number;
  export let value: string;
  export let selected: boolean = false;
</script>

<line
  x1={x}
  x2={x}
  y1={zeroY}
  y2={y}
  class={selected ? "stroke-theme-500" : "stroke-fg-muted"}
  stroke-width={selected ? 2 : 1.5}
/>
<circle
  cx={x}
  cy={y}
  r={4}
  class="stroke-surface-background stroke-[1.5px] {selected
    ? 'fill-theme-500'
    : 'fill-theme-600'}"
/>
<text
  class="stroke-surface-background text-xs font-semibold {selected
    ? 'fill-theme-500'
    : 'fill-fg-secondary'}"
  style:paint-order="stroke"
  stroke-width="3px"
  x={x + 8}
  y={y + 4}
>
  {value}
</text>
