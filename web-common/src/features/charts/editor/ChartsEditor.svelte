<script lang="ts">
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import ChartsEditorContainer from "@rilldata/web-common/features/charts/editor/ChartsEditorContainer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Editor from "../../editor/Editor.svelte";

  export let filePath: string;

  const updateFile = createRuntimeServicePutFile();
  const QUERY_DEBOUNCE_TIME = 100;

  let localContent: string | null = null;

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

  async function updateChart(content: string) {
    try {
      await $updateFile.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          path: filePath,
          blob: content,
        },
      });
    } catch (err) {
      console.error(err);
    }
  }
  const debounceUpdateChartContent = debounce(updateChart, QUERY_DEBOUNCE_TIME);

  $: hasUnsavedChanges = localContent !== null && yaml !== localContent;
</script>

<ChartsEditorContainer error={yaml?.length ? mainError : undefined}>
  <Editor
    key={filePath}
    remoteContent={yaml}
    bind:localContent
    extensions={[customYAMLwithJSONandSQL]}
    autoSave
    disableAutoSave={false}
    {hasUnsavedChanges}
    on:save={() => {
      if (localContent === null) return;
      debounceUpdateChartContent(localContent);
    }}
  />
</ChartsEditorContainer>
