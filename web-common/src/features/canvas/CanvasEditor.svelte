<script lang="ts">
  import { yaml } from "@codemirror/lang-yaml";
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { removeCanvasStore } from "./state-managers/state-managers";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";

  export let canvasName: string;
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  export let lineBasedRuntimeErrors: LineStatus[];

  let editor: EditorView;

  const instanceId = httpClient.getInstanceId();

  /** If the errors change, run the following transaction. */
  $: if (editor) setLineStatuses(lineBasedRuntimeErrors, editor);
</script>

<Editor
  bind:autoSave
  bind:editor
  onSave={(content) => {
    // Remove the canvas entity so that everything is reset to defaults next time user navigates to it
    removeCanvasStore(canvasName, instanceId);

    // Reset local persisted dashboard state for the metrics view
    if (!content?.length) {
      setLineStatuses([], editor);
    }
  }}
  {fileArtifact}
  extensions={[yaml()]}
/>
