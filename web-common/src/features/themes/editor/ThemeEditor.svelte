<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { yaml } from "@codemirror/lang-yaml";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";

  export let errors: LineStatus[];
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  let editor: EditorView;

  /** If the errors change, run the following transaction. */
  $: if (editor) setLineStatuses(errors, editor);

  /** display the main error (the first in this array) at the bottom */
  $: mainError = errors?.at(0);
</script>

<WorkspaceEditorContainer error={mainError}>
  <Editor
    bind:autoSave
    bind:editor
    onSave={(content) => {
      if (!content?.length) {
        setLineStatuses([], editor);
      }
    }}
    {fileArtifact}
    extensions={[yaml()]}
  />
</WorkspaceEditorContainer>
