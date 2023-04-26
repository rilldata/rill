<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
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

  const {
    observedNode: measuresContainerNode,
    listenToNodeResize: measuresContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  const {
    observedNode: headerContainerNode,
    listenToNodeResize: headerContainerNodeWatcher,
  } = createResizeListenerActionFactory();

  /**
   * Get the total size of an element, including margin
   * @param element
   */
  function getEltSize(element: HTMLElement, direction: "x" | "y") {
    if (!["x", "y"].includes(direction)) {
      throw new Error("direction must be 'x' or 'y'");
    }
    if (!element) return 0;
    // Get the computed style of the element
    const style = window.getComputedStyle(element);
    if (direction === "y") {
      // Get the element's height (including padding and border)
      const height = element.getBoundingClientRect().height;
      // Get the margin values
      const marginTop = parseFloat(style.marginTop);
      const marginBottom = parseFloat(style.marginBottom);
      // Calculate the total height including margin
      return height + marginTop + marginBottom;
    } else {
      const width = element.getBoundingClientRect().width;
      const marginLeft = parseFloat(style.marginLeft);
      const marginRight = parseFloat(style.marginRight);
      return width + marginLeft + marginRight;
    }
  }

  $: exploreContainerWidth = getEltSize($dashboardContainerNode, "x");
  $: exploreContainerHeight = getEltSize($dashboardContainerNode, "y");
  $: console.log("exploreContainerWidth", exploreContainerWidth);
  $: console.log(
    "$dashboardContainerNode?.offsetWidth",
    $dashboardContainerNode?.offsetWidth
  );

  $: measureContainerWidth = getEltSize($measuresContainerNode, "x");

  // $measuresContainerNode?.offsetWidth || 0;
  $: console.log("measureContainerWidth", measureContainerWidth);
  $: console.log(
    "$measuresContainerNode?.offsetWidth",
    $measuresContainerNode?.offsetWidth
  );

  // $: width = $dashboardContainerNode?.getBoundingClientRect()?.width;
  // $: console.log("width", width);
  $: headerHeight = getEltSize($headerContainerNode, "y");
  // $: console.log("headerHeight", headerHeight);
  $: targetLeaderboardContainerWidth =
    exploreContainerWidth - measureContainerWidth || 0;

  $: targetLeaderboardContainerHeight = exploreContainerHeight - headerHeight;

  $: console.log(
    "targetLeaderboardContainerHeight",
    targetLeaderboardContainerHeight
  );

  $: console.log("navigationPaddingTween", $navigationPaddingTween);
  $: console.log(
    "sum",
    targetLeaderboardContainerWidth + measureContainerWidth
  );

  $: leftSide = leftMargin
    ? leftMargin
    : `calc(${$navigationPaddingTween * 24}px + 1.25rem)`;
</script>

<section use:dashboardContainerNodeWatcher class="grid items-stretch surface">
  <div
    use:headerContainerNodeWatcher
    class="explore-header border-b mb-3"
    style:padding-left={leftSide}
    style:width={exploreContainerWidth + "px"}
  >
    <slot name="header" />
  </div>
  <div
    use:measuresContainerNodeWatcher
    class="explore-metrics mb-8"
    style:padding-left={leftSide}
  >
    <slot name="metrics" width={exploreContainerWidth} />
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
