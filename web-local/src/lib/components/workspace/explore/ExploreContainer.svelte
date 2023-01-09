<script lang="ts">
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";

  export let containerHeight = 0;
  export let gridConfig: string;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: containerHeight = $observedNode?.offsetHeight || 0;
</script>

<section
  class="grid items-stretch leaderboard-layout surface"
  style:grid-template-columns={gridConfig}
>
  <div class="explore-header">
    <slot name="header" />
  </div>
  <hr class="pb-3 pt-1 ui-divider" />
  <div
    use:listenToNodeResize
    class="explore-metrics pb-8"
    style:padding-left="calc({$navigationVisibilityTween * 24}px + 1.25rem)"
  >
    <slot name="metrics" />
  </div>
  <div class="explore-leaderboards pr-4 pb-8">
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    grid-template-rows: auto 1fr;
    height: 100vh;
    overflow-x: auto;
    overflow-y: hidden;
    grid-template-areas:
      "header header"
      "hr hr"
      "metrics leaderboards";
  }

  hr {
    grid-area: hr;
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
