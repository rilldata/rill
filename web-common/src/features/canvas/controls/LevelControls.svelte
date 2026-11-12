<script lang="ts">
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import ControlBar from "./ControlBar.svelte";

  // The viewer controls declared on the given levels, e.g. a container, or a tab group and its active tab.
  // A level is identified by its scope: [] for the canvas, [name] for a container or a tab group,
  // and [group, tab] for a tab.
  let {
    canvasEntity,
    scopes,
    maxWidth = undefined,
  }: {
    canvasEntity: CanvasEntity;
    scopes: string[][];
    // Set to center the bar like a canvas row.
    maxWidth?: number;
  } = $props();

  let enabled = $derived(canvasEntity.controls.enabled);
  let all = $derived(canvasEntity.controls.all);
  let keys = $derived(new Set(scopes.map((s) => s.join("\u0000"))));
  let controls = $derived(
    $all.filter(
      (c) =>
        c.level !== "item" && !c.locked && keys.has(c.scope.join("\u0000")),
    ),
  );
</script>

{#if $enabled && controls.length}
  <div
    class="level-controls"
    style:max-width={maxWidth ? `${maxWidth}px` : undefined}
  >
    <ControlBar {controls} {canvasEntity} />
  </div>
{/if}

<style lang="postcss">
  .level-controls {
    @apply pointer-events-auto w-full px-2 py-1;
  }
</style>
