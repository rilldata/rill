<script lang="ts">
  import ControlBar from "@rilldata/web-common/features/canvas/controls/ControlBar.svelte";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

  // PDF-capture render of the viewer controls of a level (a container, a tab group or a tab):
  // a band of text above the level's rows, e.g. "Metric: Impressions", so the PDF says which
  // choices its charts reflect. Renders nothing for a level without visible controls.
  // With `inline`, the text renders without the band, for use inside another captured block.
  let {
    canvasEntity,
    scopes,
    maxWidth = undefined,
    inline = false,
  }: {
    canvasEntity: CanvasEntity;
    // The levels to show, by scope: [] for the canvas, [name] for a container or a tab group,
    // [group, tab] for a tab.
    scopes: string[][];
    maxWidth?: number;
    inline?: boolean;
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
  {#if inline}
    <ControlBar {controls} {canvasEntity} />
  {:else}
    <section
      id={`pdf-controls-${scopes[scopes.length - 1]?.join("-") ?? ""}`}
      class="pdf-controls-row w-full h-fit relative px-2 pt-2"
      style:max-width="{maxWidth}px"
    >
      <ControlBar {controls} {canvasEntity} />
    </section>
  {/if}
{/if}
