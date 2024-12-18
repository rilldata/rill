<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import ComponentsEditorContainer from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditorContainer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Editor from "../../editor/Editor.svelte";

  export let filePath: string;

  let editor: EditorView;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: ({ autoSave, remoteContent } = fileArtifact);

  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    $allErrors,
    $remoteContent ?? "",
  );
  /** display the main error (the first in this array) at the bottom */
  $: mainError = lineBasedRuntimeErrors?.at(0);
</script>

<ComponentsEditorContainer
  error={$remoteContent?.length ? mainError : undefined}
>
  <Editor
    {fileArtifact}
    extensions={[customYAMLwithJSONandSQL]}
    bind:editor
    bind:autoSave={$autoSave}
  />
</ComponentsEditorContainer>
