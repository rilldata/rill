<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import APITestPanel from "./APITestPanel.svelte";
  import type { Arg } from "./types";

  export let apiName: string;
  export let errors: LineStatus[];
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  export let isReconciling: boolean;
  export let host: string;
  export let instanceId: string;
  export let args: Arg[];

  let editor: EditorView;

  $: if (editor) setLineStatuses(errors, editor);
  $: mainError = errors?.at(0)?.message;
  $: hasErrors = errors.length > 0;
</script>

<div class="flex flex-col h-full overflow-hidden">
  <div class="editor-panel">
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
        extensions={[customYAMLwithJSONandSQL]}
      />
    </WorkspaceEditorContainer>
  </div>

  <APITestPanel
    {apiName}
    {hasErrors}
    {isReconciling}
    {host}
    {instanceId}
    bind:args
  />
</div>

<style lang="postcss">
  .editor-panel {
    @apply flex-1 overflow-hidden min-h-0;
  }
</style>
