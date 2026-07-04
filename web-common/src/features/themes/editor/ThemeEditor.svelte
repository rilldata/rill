<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import { yaml } from "@codemirror/lang-yaml";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { mapParseErrorToLine } from "@rilldata/web-common/features/metrics-views/errors";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";

  export let parseError: V1ParseError | undefined = undefined;
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  $: ({ remoteContent } = fileArtifact);

  let editor: EditorView;

  /** If the parse error changes, update the editor gutter. */
  $: lineStatus = mapParseErrorToLine(parseError, $remoteContent ?? "");
  $: if (editor) setLineStatuses(lineStatus ? [lineStatus] : [], editor);
</script>

<WorkspaceEditorContainer error={parseError?.message}>
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
