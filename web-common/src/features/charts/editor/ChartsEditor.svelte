<script lang="ts">
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import ChartsEditorContainer from "@rilldata/web-common/features/charts/editor/ChartsEditorContainer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import Editor from "../../editor/Editor.svelte";
  import type { EditorView } from "@codemirror/view";

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

<ChartsEditorContainer error={$remoteContent?.length ? mainError : undefined}>
  <Editor
    {fileArtifact}
    extensions={[customYAMLwithJSONandSQL]}
    bind:editor
    bind:autoSave={$autoSave}
  />
</ChartsEditorContainer>
