<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
  import { getEltSize } from "@rilldata/web-common/features/dashboards/get-element-size";
  import type { Tweened } from "svelte/motion";

  export let exploreContainerWidth;

  export let leftMargin: string = undefined;

  // the navigationPaddingTween is a tweened value that is used
  // to animate the extra padding that needs to be added to the
  // dashboard container when the navigation pane is collapsed
  const navigationPaddingTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const {
    observedNode: dashboardContainerNode,
    listenToNodeResize: dashboardContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  $: exploreContainerWidth = getEltSize($dashboardContainerNode, "x");

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationPaddingTween * 24}px + 1.25rem)`;
</script>

<section use:dashboardContainerNodeWatcher class="flex flex-col gap-y-1">
  <div
    class="explore-header border-b mb-3"
    style:padding-left={leftSide}
    style:width={"100%"}
  >
    <slot name="header" />
  </div>
  <div
    class="explore-content flex flex-row gap-x-1"
    style:padding-left={leftSide}
  >
    <div class="explore-metrics mb-8 flex-none">
      <slot name="metrics" />
    </div>
    <div class="explore-leaderboards px-4 mb-8 grow">
      <slot name="leaderboards" />
    </div>
  </div>
</section>

<style>
  section {
    height: 100vh;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .explore-header {
    grid-area: header;
  }
  .explore-content {
    height: 100%;
    overflow: hidden;
  }
  .explore-metrics {
    overflow-y: scroll;
  }

  .explore-leaderboards {
    overflow-y: hidden;
  }
</style>
