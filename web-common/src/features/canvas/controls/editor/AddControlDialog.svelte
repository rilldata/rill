<script lang="ts">
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import { applyCanvasEdit } from "../../inspector/BlockEditor.svelte";
  import type { YAMLPath } from "../../layout-util";
  import {
    addControl,
    type ControlDraft,
    type ControlLevelRef,
  } from "../control-edit";
  import ControlDialog from "./ControlDialog.svelte";

  // The dialog for a new viewer control, opened from the canvas: the button at the top of
  // the canvas or after a control bar, or a block's or widget's menu. The inspector's
  // "Viewer controls" sections open the same form for editing.
  let {
    canvas,
    levels,
    levelPath,
    metricsViewName = undefined,
    onClose,
  }: {
    canvas: CanvasEntity;
    // The levels the control can be declared on, nearest first. With one level, the choice is hidden.
    levels: ControlLevelRef[];
    // The level chosen when the dialog opens.
    levelPath: YAMLPath;
    // For a widget: its metrics view, whose fields the dialog offers.
    metricsViewName?: string;
    onClose: () => void;
  } = $props();

  async function save(draft: ControlDraft, path: YAMLPath) {
    onClose();
    if (!canvas.fileArtifact) return;
    await applyCanvasEdit(canvas.fileArtifact, true, (doc) => {
      addControl(doc, path, draft);
    });
  }
</script>

<ControlDialog
  {canvas}
  {levels}
  {levelPath}
  isNew
  {metricsViewName}
  onSave={save}
  onCancel={onClose}
/>
