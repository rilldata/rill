<script lang="ts">
  import { createResizeListenerActionFactory } from "../actions/create-resize-listener-factory";

  export let side: "left" | "right";

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
</script>

<div
  use:listenToNodeResize
  class="sticky top-0 bg-white z-10 px-4 flex flex-row items-center gap-x-2 justify-{side ===
  'left'
    ? 'start'
    : 'end'}"
  style:height="var(--header-height)"
  style="padding-{side === 'left' ? 'right' : 'left'}: 40px;"
>
  <slot width={$observedNode?.getBoundingClientRect()?.width} />
</div>
