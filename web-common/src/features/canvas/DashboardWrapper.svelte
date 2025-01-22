<script lang="ts">
  import * as defaults from "./constants";

  export let width: number;
  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let bgColor = "bg-white";
</script>

<div
  class="dashboard-theme-boundary size-full flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="dashboard-wrapper {bgColor} min-h-full"
    style:max-width="{defaults.DEFAULT_DASHBOARD_WIDTH}px "
    style:height="{height}px"
  >
    <div role="presentation" class="size-full relative" on:click|self>
      <div
        class="dash pointer-events-none"
        role="presentation"
        style:width="{width}px"
        style:height="{height}px"
      >
        <slot />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .dashboard-wrapper {
    width: 100%;
    height: 100%;
    position: relative;
    user-select: none;
    margin: 0;
    pointer-events: auto;
  }

  .dash {
    position: absolute;
    touch-action: none;
  }
</style>
