<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { yaml } from "@codemirror/lang-yaml";

  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  let editor: EditorView;
</script>

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
