<script lang="ts">
  export let height: number;
  export let top: number;
  export let left: number;
  export let orientation: "vertical" | "horizontal" = "vertical";
</script>

<div
  class="ghost-line"
  style:height="{orientation === 'vertical' ? height : '2'}px"
  style:width="{orientation === 'horizontal' ? height : '2'}px"
  style:top="{top}px"
  style:left="{left}px"
  data-orientation={orientation}
/>

<style>
  .ghost-line {
    position: absolute;
    background-color: #4299e1;
    pointer-events: none;
    transition: all 0.1s ease;
    z-index: 1000;
    box-shadow: 0 0 12px rgba(66, 153, 225, 0.8);
    border-radius: 3px;
    opacity: 0.8;
  }
</style>
