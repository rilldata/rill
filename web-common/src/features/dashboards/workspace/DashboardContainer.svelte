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

  const {
    observedNode: dashboardContainerNode,
    listenToNodeResize: dashboardContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  const {
    observedNode: measuresContainerNode,
    listenToNodeResize: measuresContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  const {
    observedNode: headerContainerNode,
    listenToNodeResize: headerContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  $: exploreContainerWidth = $dashboardContainerNode?.offsetWidth || 0;
  // $: console.log("exploreContainerWidth", exploreContainerWidth);

  $: measureContainerWidth = $measuresContainerNode?.offsetWidth || 0;
  // $: console.log("measureContainerWidth", measureContainerWidth);

  $: width = $dashboardContainerNode?.getBoundingClientRect()?.width;
  // $: console.log("width", width);

  $: targetLeaderboardContainerWidth =
    exploreContainerWidth - measureContainerWidth || 0;

  $: targetLeaderboardContainerHeight =
    $dashboardContainerNode?.getBoundingClientRect()?.height -
      $headerContainerNode?.getBoundingClientRect()?.height || 0;

  $: console.log(
    "targetLeaderboardContainerHeight",
    targetLeaderboardContainerHeight
  );

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationVisibilityTween * 24}px + 1.25rem)`;
</script>

<section use:dashboardContainerNodeWatcher class="grid items-stretch surface">
  <div
    use:headerContainerNodeWatcher
    class="explore-header border-b mb-3"
    style:padding-left={leftSide}
    style:width={width + "px"}
  >
    <slot name="header" />
  </div>
  <div
    use:measuresContainerNodeWatcher
    class="explore-metrics mb-8"
    style:padding-left={leftSide}
  >
    <slot
      name="metrics"
      width={$dashboardContainerNode?.getBoundingClientRect()?.width}
    />
  </div>
  <div
    class="explore-leaderboards px-4"
    style={`height:${targetLeaderboardContainerHeight}px; width:${targetLeaderboardContainerWidth}px`}
  >
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    grid-template-rows: auto auto 1fr;
    grid-template-columns: min-content 1fr;
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
