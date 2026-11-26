<script lang="ts">
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Plus } from "lucide-svelte";
  import type { YAMLPath } from "../../layout-util";
  import type { ControlLevelRef } from "../control-edit";
  import AddControlDialog from "./AddControlDialog.svelte";

  // The editor's on-canvas button for adding a viewer control to a level. It reads
  // "Add viewer control" where the level has none yet, and "Add control" after a control bar.
  let {
    canvas,
    levels,
    levelPath,
    compact = false,
  }: {
    canvas: CanvasEntity;
    levels: ControlLevelRef[];
    levelPath: YAMLPath;
    compact?: boolean;
  } = $props();

  let open = $state(false);
</script>

<!-- The canvas behind the button selects and drags on mouse events, so they stop here. -->
<button
  type="button"
  class="add-control"
  onmousedown={(e) => e.stopPropagation()}
  onclick={(e) => {
    e.stopPropagation();
    open = true;
  }}
>
  <Plus size="12px" />
  {compact
    ? m.canvas_viewer_controls_add()
    : m.canvas_viewer_controls_add_viewer()}
</button>

{#if open}
  <AddControlDialog
    {canvas}
    {levels}
    {levelPath}
    onClose={() => (open = false)}
  />
{/if}

<style lang="postcss">
  .add-control {
    @apply flex h-6 flex-none items-center gap-x-1 whitespace-nowrap rounded-sm px-2;
    @apply border border-dashed text-xs text-fg-secondary;
  }

  .add-control:hover {
    @apply border-primary-400 bg-surface-hover text-fg-primary;
  }
</style>
