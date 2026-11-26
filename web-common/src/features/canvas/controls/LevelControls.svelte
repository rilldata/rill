<script lang="ts">
  import { controlsForScopes } from "./controls";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import type { YAMLPath } from "../layout-util";
  import type { ControlLevelRef } from "./control-edit";
  import ControlBar from "./ControlBar.svelte";
  import AddControlButton from "./editor/AddControlButton.svelte";

  // The viewer controls declared on the given levels, e.g. a container, or a tab group and its active tab.
  // A level is identified by its scope: [] for the canvas, [name] for a container or a tab group,
  // and [group, tab] for a tab.
  let {
    canvasEntity,
    scopes,
    maxWidth = undefined,
    add = undefined,
  }: {
    canvasEntity: CanvasEntity;
    scopes: string[][];
    // Set to center the bar like a canvas row.
    maxWidth?: number;
    // In the editor: where the add button after the bar adds a control. The levels are the
    // ones the dialog offers, and `whenEmpty` keeps the button, as "Add viewer control", on a
    // level without controls (the canvas; blocks offer it in their menus instead).
    add?: {
      levels: ControlLevelRef[];
      levelPath: YAMLPath;
      whenEmpty?: boolean;
    };
  } = $props();

  let enabled = $derived(canvasEntity.controls.enabled);
  let all = $derived(canvasEntity.controls.all);
  let controls = $derived(
    controlsForScopes($all, scopes).filter((c) => !c.locked),
  );
</script>

{#if $enabled && (controls.length || add?.whenEmpty)}
  <div
    class="level-controls"
    style:max-width={maxWidth ? `${maxWidth}px` : undefined}
  >
    {#if controls.length}
      <div class="min-w-0 flex-1">
        <ControlBar {controls} {canvasEntity} />
      </div>
    {/if}
    {#if add}
      <AddControlButton
        canvas={canvasEntity}
        levels={add.levels}
        levelPath={add.levelPath}
        compact={controls.length > 0}
      />
    {/if}
  </div>
{/if}

<style lang="postcss">
  .level-controls {
    @apply pointer-events-auto flex w-full items-center gap-x-3 px-2 py-1;
  }
</style>
