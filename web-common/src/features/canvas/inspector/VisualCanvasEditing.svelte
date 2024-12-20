<script lang="ts">
  import { replaceState } from "$app/navigation";
  import ComponentsEditor from "@rilldata/web-common/features/canvas/inspector/ComponentsEditor.svelte";
  import PageEditor from "@rilldata/web-common/features/canvas/inspector/PageEditor.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { parseDocument } from "yaml";

  export let fileArtifact: FileArtifact;

  const { validSpecStore, canvasStore } = getCanvasStateManagers();

  $: ({ editorContent, remoteContent, saveContent, path } = fileArtifact);

  $: parsedDocument = parseDocument($editorContent ?? $remoteContent ?? "");
  $: selectedComponentIndex = $canvasStore.selectedComponentIndex;

  $: selectedComponentName =
    selectedComponentIndex !== null
      ? $validSpecStore?.data?.items?.[selectedComponentIndex]?.component
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

    await saveContent(parsedDocument.toString());
  }

  function killState() {
    replaceState(window.location.origin + window.location.pathname, {});
  }
</script>

<Inspector filePath={path}>
  {#if selectedComponentName}
    <ComponentsEditor {fileArtifact} {selectedComponentName} />
  {:else}
    <PageEditor {fileArtifact} {updateProperties} />
  {/if}
</Inspector>
