<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { clearExploreViewState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
  import { clearLastVisitedState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/last-visited-state";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { yaml } from "@codemirror/lang-yaml";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";

  export let exploreName: string;
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  export let lineBasedRuntimeErrors: LineStatus[];

  let editor: EditorView;

  /** If the errors change, run the following transaction. */
  $: if (editor) setLineStatuses(lineBasedRuntimeErrors, editor);
</script>

<Editor
  bind:autoSave
  bind:editor
  onSave={(content) => {
    // Remove the explorer entity so that everything is reset to defaults next time user navigates to it
    metricsExplorerStore.remove(exploreName);
    clearExploreViewState(exploreName, undefined);
    clearLastVisitedState(exploreName);

    if (!content?.length) {
      setLineStatuses([], editor);
    }
  }}
  {fileArtifact}
  extensions={[yaml()]}
/>
