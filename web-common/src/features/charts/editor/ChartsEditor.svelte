<script lang="ts">
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import ChartsEditorContainer from "@rilldata/web-common/features/charts/editor/ChartsEditorContainer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Editor from "../../editor/Editor.svelte";

  export let filePath: string;
  export let autoSave: boolean;

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, {
    path: filePath,
  });
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  // get the yaml blob from the file.
  $: yaml = $fileQuery.data?.blob || "";

  const queryClient = useQueryClient();
  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  $: lineBasedRuntimeErrors = mapParseErrorsToLines($allErrors, yaml);
  /** display the main error (the first in this array) at the bottom */
  $: mainError = lineBasedRuntimeErrors?.at(0);
</script>

<ChartsEditorContainer error={yaml?.length ? mainError : undefined}>
  <Editor
    {fileArtifact}
    extensions={[customYAMLwithJSONandSQL]}
    bind:autoSave
    disableAutoSave={false}
  />
</ChartsEditorContainer>
