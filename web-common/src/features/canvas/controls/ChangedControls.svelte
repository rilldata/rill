<script lang="ts">
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // Tells viewers that some controls are off their defaults, and lets them go back.
  let {
    canvasEntity,
    maxWidth = undefined,
  }: { canvasEntity: CanvasEntity; maxWidth?: number } = $props();

  let changed = $derived(canvasEntity.controls.changed);
</script>

{#if $changed.length}
  <div
    class="changed-controls"
    style:max-width={maxWidth ? `${maxWidth}px` : undefined}
  >
    <span>{m.canvas_controls_changed()}</span>
    <button type="button" onclick={() => canvasEntity.controls.resetAll()}>
      {m.canvas_controls_reset_all()}
    </button>
  </div>
{/if}

<style lang="postcss">
  .changed-controls {
    @apply pointer-events-auto flex w-full items-center justify-end gap-x-2 px-2 pt-1;
    @apply text-xs text-fg-secondary;
  }

  button {
    @apply text-primary-600 font-medium;
  }

  button:hover {
    @apply underline;
  }
</style>
