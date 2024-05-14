<script lang="ts">
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { V1ParseError } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import ChartsEditorContainer from "../charts/editor/ChartsEditorContainer.svelte";
  import Editor from "../editor/Editor.svelte";
  import { FileExtensionToEditorExtension } from "../editor/getExtensionsForFile";

  const dispatch = createEventDispatcher();

  export let yaml: string;
  export let errors: V1ParseError[] = [];
  export let filePath: string;

  const QUERY_DEBOUNCE_TIME = 300;

  function updateChart(content: string) {
    dispatch("update", content);
  }
  const debounceUpdateChartContent = debounce(updateChart, QUERY_DEBOUNCE_TIME);

  let localContent: string;

  $: hasUnsavedChanges = yaml !== localContent;
</script>

<ChartsEditorContainer error={errors[0]}>
  <Editor
    blob={yaml}
    key={filePath}
    bind:latest={localContent}
    extensions={FileExtensionToEditorExtension[".yaml"]}
    autoSave
    disableAutoSave={false}
    showSaveControls={false}
    {hasUnsavedChanges}
    on:save={() => debounceUpdateChartContent(localContent)}
  />
</ChartsEditorContainer>
