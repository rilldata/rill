<script lang="ts">
  import { replaceState } from "$app/navigation";
  import ComponentsEditor from "@rilldata/web-common/features/canvas/inspector/ComponentsEditor.svelte";
  import PageEditor from "@rilldata/web-common/features/canvas/inspector/PageEditor.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { parseDocument } from "yaml";

  const { canvasStore, fileArtifact } = getCanvasStateManagers();

  $: ({ localContent, remoteContent, saveContent, path } = $fileArtifact);

  $: parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
  $: selectedComponentIndex = $canvasStore.selectedComponentIndex;

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
  {#if selectedComponentIndex !== null}
    <ComponentsEditor />
  {:else}
    <PageEditor {updateProperties} />
  {/if}
</Inspector>
