<script lang="ts">
  import {
    assetVisibilityTween,
    inspectorVisibilityTween,
    layout,
    modelPreviewVisibilityTween,
    SIDE_PAD,
  } from "../../../application-state-stores/layout-store";
  import { drag } from "../../../drag";
  import Portal from "../../Portal.svelte";
  let innerHeight;
</script>

<Portal target=".body">
  <div
    class="fixed drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center ml-2 mr-2"
    style:bottom="{(1 - $modelPreviewVisibilityTween) *
      $layout.modelPreviewHeight}px"
    style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth + 16}px"
    style:right="{(1 - $inspectorVisibilityTween) * $layout.inspectorWidth +
      16}px"
    style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
    style:padding-right="{$inspectorVisibilityTween * SIDE_PAD}px"
    use:drag={{
      minSize: 200,
      maxSize: innerHeight - 200,
      side: "modelPreviewHeight",
      orientation: "vertical",
      reverse: true,
    }}
  >
    <div class="border-t border-gray-300" />
    <div class="absolute right-1/2 left-1/2 top-1/2 bottom-1/2">
      <div
        class="border-gray-400 border bg-white rounded h-1 w-8 absolute -translate-y-1/2"
      />
    </div>
  </div>
</Portal>
