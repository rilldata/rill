<script lang="ts">
  import { createResizeListenerActionFactory } from "../actions/create-resize-listener-factory";

  export let side: "left" | "right";

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: width = $observedNode?.getBoundingClientRect()?.width;
</script>

<div
  use:listenToNodeResize
  class=" px-4 flex flex-row items-center gap-x-2 justify-{side === 'left'
    ? 'start'
    : 'end'}"
  style:height="var(--header-height)"
>
  {#if width}
    <slot {width} />
  {/if}
</div>
