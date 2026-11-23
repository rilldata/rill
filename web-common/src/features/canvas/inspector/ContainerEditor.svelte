<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Ungroup } from "lucide-svelte";
  import { containerRowsPath } from "../layout-util";
  import { unwrapRows } from "../stores/block-edit";
  import type { CanvasEntity } from "../stores/canvas-entity";
  import type { RowsBlock } from "../stores/layout-blocks";
  import BlockEditor, { applyCanvasEdit } from "./BlockEditor.svelte";

  // The inspector of a container: the shared block inspector, plus Ungroup.
  let {
    container,
    blockIndex,
    canvas,
    fileArtifact,
    autoSave,
  }: {
    container: RowsBlock;
    blockIndex: number;
    canvas: CanvasEntity;
    fileArtifact: FileArtifact;
    autoSave: boolean;
  } = $props();

  // Replace the container with its rows, which keep its condition (see unwrapRows).
  async function ungroup() {
    await applyCanvasEdit(fileArtifact, autoSave, (doc) => {
      if (unwrapRows(doc, blockIndex, containerRowsPath(blockIndex)) >= 0) {
        canvas.setSelectedBlock(null);
      }
    });
  }
</script>

<BlockEditor
  kind="container"
  name={container.name}
  {blockIndex}
  {canvas}
  {fileArtifact}
  {autoSave}
>
  {#snippet actions()}
    <Button type="secondary" onClick={ungroup}>
      <Ungroup size="14px" />
      {m.canvas_container_ungroup()}
    </Button>
  {/snippet}
</BlockEditor>
