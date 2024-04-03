<script lang="ts">
  import Resizer from "../Resizer.svelte";
  import { workspaces } from "./workspace-stores";
  import { page } from "$app/stores";

  let resizing = false;

  $: context = $page.url.pathname;
  $: workspace = workspaces.get(context);
  $: width = workspace.inspector.width;
  $: visible = workspace.inspector.visible;
</script>

<div
  class="inspector-wrapper"
  class:closed={!$visible}
  class:resizing
  style:width="{$width}px"
>
  <Resizer
    direction="EW"
    side="left"
    min={300}
    max={500}
    bind:dimension={$width}
    bind:resizing
  />

  <div class="inner" style:width="{$width}px">
    <slot />
  </div>
</div>

<style lang="postcss">
  .inspector-wrapper {
    will-change: width;
    @apply h-full flex-none relative;
    @apply border-l border-gray-200 bg-white;
    @apply overflow-y-auto overflow-x-hidden;
  }

  .inner {
    will-change: width;
    @apply h-fit;
  }

  .inspector-wrapper:not(.resizing) {
    transition-property: width;
    transition-duration: 600ms;
    transition-timing-function: cubic-bezier(0.22, 1, 0.36, 1);
  }

  .closed {
    width: 0px !important;
  }
</style>
