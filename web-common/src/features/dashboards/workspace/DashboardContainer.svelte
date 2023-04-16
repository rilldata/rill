<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";

  export let exploreContainerWidth;
  export let width;

  export let leftMargin: string = undefined;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: exploreContainerWidth = $observedNode?.offsetWidth || 0;

  $: width = $observedNode?.getBoundingClientRect()?.width;

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationVisibilityTween * 24}px + 1.25rem)`;
</script>

<section
  use:listenToNodeResize
  class="grid items-stretch surface"
  style:grid-template-columns="max-content minmax(355px, auto)"
>
  <div class="explore-header border-b mb-3" style:padding-left={leftSide}>
    <slot name="header" />
  </div>
  <div class="explore-metrics mb-8" style:padding-left={leftSide}>
    <slot
      name="metrics"
      width={$observedNode?.getBoundingClientRect()?.width}
    />
  </div>
  <div class="explore-leaderboards pr-4 pb-8">
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    grid-template-rows: auto auto 1fr;
    height: 100vh;
    overflow-x: auto;
    overflow-y: hidden;
    grid-template-areas:
      "header header"
      "metrics leaderboards";
  }

  .explore-header {
    grid-area: header;
  }
  .explore-metrics {
    grid-area: metrics;
    overflow-y: auto;
  }
  .explore-leaderboards {
    grid-area: leaderboards;
  }
</style>
