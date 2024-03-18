<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { customYAMLWithJSON } from "@rilldata/web-common/components/editor/presets/yamlWithJson";
  import ChartsEditorContainer from "@rilldata/web-common/features/charts/editor/ChartsEditorContainer.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let chartName: string;

  const updateFile = createRuntimeServicePutFile();
  const QUERY_DEBOUNCE_TIME = 100;

  let view: EditorView;
  let editor: YAMLEditor;

  $: filePath = getFilePathFromNameAndType(chartName, EntityType.Chart);
  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);

  // get the yaml blob from the file.
  $: yaml = $fileQuery.data?.blob || "";

  const queryClient = useQueryClient();
  $: allErrors = getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    getFilePathFromNameAndType(chartName, EntityType.Chart),
  );

  $: lineBasedRuntimeErrors = mapParseErrorsToLines($allErrors, yaml);
  /** display the main error (the first in this array) at the bottom */
  $: mainError = lineBasedRuntimeErrors?.at(0);

  async function updateChart(content: string) {
    try {
      await $updateFile.mutateAsync({
        instanceId: $runtime.instanceId,
        path: getFileAPIPathFromNameAndType(chartName, EntityType.Chart),
        data: {
          blob: content,
        },
      });
    } catch (err) {
      console.error(err);
    }
  }
  const debounceUpdateChartContent = debounce(updateChart, QUERY_DEBOUNCE_TIME);
</script>

<ChartsEditorContainer error={yaml?.length ? mainError : undefined}>
  <YAMLEditor
    bind:this={editor}
    bind:view
    content={yaml}
    extensions={[customYAMLWithJSON]}
    on:update={(e) => debounceUpdateChartContent(e.detail.content)}
  />
</ChartsEditorContainer>
