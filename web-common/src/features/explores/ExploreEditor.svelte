<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { createPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";
  import { yaml } from "@codemirror/lang-yaml";
  import MetricsEditorContainer from "../metrics-views/editor/MetricsEditorContainer.svelte";

  export let exploreName: string;
  export let allErrors: V1ParseError[];
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  $: ({ remoteContent } = fileArtifact);

  let editor: EditorView;

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    allErrors,
    $remoteContent ?? "",
  );
  /** display the main error (the first in this array) at the bottom */
  $: mainError = lineBasedRuntimeErrors?.at(0);

  /** If the errors change, run the following transaction. */
  $: if (editor) setLineStatuses(lineBasedRuntimeErrors, editor);
</script>

<MetricsEditorContainer error={$remoteContent ? mainError : undefined}>
  <Editor
    bind:autoSave
    bind:editor
    onSave={(content) => {
      // Remove the explorer entity so that everything is reset to defaults next time user navigates to it
      metricsExplorerStore.remove(exploreName);
      // Reset local persisted dashboard state for the metrics view
      createPersistentDashboardStore(exploreName).reset();

      if (!content?.length) {
        setLineStatuses([], editor);
      }
    }}
    {fileArtifact}
    extensions={[yaml()]}
  />
</MetricsEditorContainer>
