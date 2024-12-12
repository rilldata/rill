<script lang="ts">
  export let width: number;
  export let height: number;
  export let scale: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-slate-200";
</script>

<div
  class="dashboard-theme-boundary size-full bg-gray-100 flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="wrapper {color} max-w-[1440px] min-h-full"
    style:height="{height}px"
  >
    <div role="presentation" class="size-full relative" on:click|self>
      <div
        class="dash pointer-events-none"
        role="presentation"
        style:width="{width}px"
        style:height="{height}px"
        style:transform="scale({scale})"
      >
        <slot />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    width: 100%;
    height: 100%;
    position: relative;
    user-select: none;
    margin: 0;
    pointer-events: auto;
  }

  .dash {
    transform-origin: top left;
    position: absolute;
    touch-action: none;
  }
</style>
