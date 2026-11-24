<script lang="ts">
  import { replaceState } from "$app/navigation";
  import ComponentsEditor from "@rilldata/web-common/features/canvas/inspector/ComponentsEditor.svelte";
  import ContainerEditor from "@rilldata/web-common/features/canvas/inspector/ContainerEditor.svelte";
  import PageEditor from "@rilldata/web-common/features/canvas/inspector/PageEditor.svelte";
  import TabGroupEditor from "@rilldata/web-common/features/canvas/inspector/TabGroupEditor.svelte";
  import { blockIndexOf } from "@rilldata/web-common/features/canvas/stores/block-edit";
  import { layoutBlockName } from "@rilldata/web-common/features/canvas/stores/layout-blocks";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { parseDocument } from "yaml";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  export let canvasName: string;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);

  $: ({ canvasEntity } = getCanvasStore(canvasName, instanceId));
  $: ({ selectedComponent, componentsStore, selectedBlock, layout } =
    canvasEntity);

  // Resolve the selected block (a tab group or a container) to its layout block, for its
  // instance.
  $: block =
    $selectedBlock != null
      ? $layout.find((b) => layoutBlockName(b) === $selectedBlock)
      : undefined;

  $: ({ editorContent, updateEditorContent, saveLocalContent, path } =
    fileArtifact);

  $: parsedDocument = parseDocument($editorContent ?? "");

  // The selected block's top-level index, looked up by name in the latest YAML. The layout
  // catches up with an edit only once the spec is reconciled, so until then its index can
  // belong to another block, and the inspector's edits would go to that block.
  $: blockIndex =
    $selectedBlock != null ? blockIndexOf(parsedDocument, $selectedBlock) : -1;

  $: components = $componentsStore;
  $: component = components.get($selectedComponent ?? "");

  $: isCustomChartComponent = component?.type === "custom_chart";

  async function updateProperties(
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
  ) {
    Object.entries(newRecord).forEach(([property, value]) => {
      if (!value) {
        parsedDocument.delete(property);
      } else {
        parsedDocument.set(property, value);
      }
    });

    if (removeProperties) {
      removeProperties.forEach((prop) => {
        try {
          if (Array.isArray(prop)) {
            parsedDocument.deleteIn(prop);
          } else {
            parsedDocument.delete(prop);
          }
        } catch (e) {
          console.warn(e);
        }
      });
    }

    killState();
    updateEditorContent(parsedDocument.toString(), false, autoSave);
    await saveLocalContent();
  }

  function killState() {
    replaceState(window.location.origin + window.location.pathname, {});
  }
</script>

<Inspector
  minWidth={320}
  filePath={path}
  maxWidth={isCustomChartComponent ? 600 : 420}
>
  {#if component}
    <ComponentsEditor {component} />
  {:else if block?.kind === "tab-group" && blockIndex >= 0}
    <TabGroupEditor
      group={block.group}
      {blockIndex}
      canvas={canvasEntity}
      {fileArtifact}
      {autoSave}
    />
  {:else if block?.kind === "container" && blockIndex >= 0}
    <ContainerEditor
      container={block.container}
      {blockIndex}
      canvas={canvasEntity}
      {fileArtifact}
      {autoSave}
    />
  {:else}
    <PageEditor {canvasName} {fileArtifact} {updateProperties} />
  {/if}
</Inspector>
