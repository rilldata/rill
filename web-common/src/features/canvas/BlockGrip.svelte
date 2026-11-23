<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { GripVertical } from "lucide-svelte";
  import type { BlockKind } from "./stores/block-edit";

  // The handle that drags a whole block (a tab group or a container) to another position
  // on the canvas. The parent starts the drag on mousedown; `element` lets it tell a press
  // on the grip apart from one on its other controls.
  let {
    kind,
    class: className = "",
    element = $bindable(),
  }: {
    kind: BlockKind;
    class?: string;
    element?: HTMLDivElement;
  } = $props();

  const title = $derived(
    kind === "tab-group"
      ? m.canvas_block_drag_tab_group()
      : m.canvas_block_drag_container(),
  );
</script>

<div
  bind:this={element}
  class={["grip", className]}
  {title}
  role="presentation"
>
  <GripVertical size="14px" />
</div>

<style lang="postcss">
  .grip {
    @apply flex flex-none items-center cursor-grab active:cursor-grabbing;
    @apply text-fg-secondary hover:text-fg-primary;
  }
</style>
