<script lang="ts">
  export let width: number | undefined;
  export let height: number | undefined;
  export let x: number | undefined;
  export let y: number | undefined;
  export let content: string | undefined;
</script>

<div
  class="grid-stack-item"
  data-gs-w={width}
  data-gs-h={height}
  data-gs-x={x}
  data-gs-y={y}
>
  <div class="grid-stack-item-content">
    {content}
  </div>
</div>

<style lang="postcss">
  .grid-stack-item-content {
    background-color: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }
</style>
