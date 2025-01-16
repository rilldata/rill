<script lang="ts">
  export let height: number;
  export let top: number;
  export let left: number;
  export let orientation: "horizontal" | "vertical";
  export let width: number | undefined = undefined;
</script>

<div
  class="absolute bg-primary-500"
  style="
    height: {height}px;
    top: {top}px;
    left: {left}px;
    width: {orientation === 'vertical' ? '2px' : width ? `${width}px` : 'auto'};
    {orientation === 'horizontal' && !width ? 'right: 0;' : ''}
  "
/>
