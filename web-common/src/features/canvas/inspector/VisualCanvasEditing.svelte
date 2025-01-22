<script lang="ts">
  import { replaceState } from "$app/navigation";
  import ComponentsEditor from "@rilldata/web-common/features/canvas/inspector/ComponentsEditor.svelte";
  import PageEditor from "@rilldata/web-common/features/canvas/inspector/PageEditor.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { parseDocument } from "yaml";

  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  const { canvasEntity } = getCanvasStateManagers();
  const { canvasSpec } = canvasEntity.spec;

  $: ({ editorContent, updateEditorContent, saveLocalContent, path } =
    fileArtifact);

  $: parsedDocument = parseDocument($editorContent ?? "");
  $: selectedComponentIndex = canvasEntity.selectedComponentIndex;

  $: selectedComponentName =
    $selectedComponentIndex !== null
      ? $canvasSpec?.items?.[$selectedComponentIndex]?.component
      : null;

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

<Inspector minWidth={320} filePath={path}>
  {#if selectedComponentName}
    <ComponentsEditor {fileArtifact} {selectedComponentName} />
  {:else}
    <PageEditor {fileArtifact} {updateProperties} />
  {/if}
</Inspector>
