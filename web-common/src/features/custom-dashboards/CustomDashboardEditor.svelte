<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let customDashboardName: string;

  const updateFile = createRuntimeServicePutFile();
  const QUERY_DEBOUNCE_TIME = 400;

  let view: EditorView;
  let editor: YAMLEditor;

  $: filePath = getFilePathFromNameAndType(
    customDashboardName,
    EntityType.Dashboard,
  );
  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);

  $: yaml = $fileQuery.data?.blob || "";

  async function updateChart(content: string) {
    try {
      await $updateFile.mutateAsync({
        instanceId: $runtime.instanceId,
        path: getFileAPIPathFromNameAndType(
          customDashboardName,
          EntityType.Dashboard,
        ),
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

<YAMLEditor
  bind:this={editor}
  bind:view
  content={yaml ?? ""}
  on:update={(e) => debounceUpdateChartContent(e.detail.content)}
/>
