<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { parse } from "yaml";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFileAndReconcile,
    getRuntimeServiceGetFileQueryKey,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import YAMLEditor from "../YAMLEditor.svelte";
  import YamlEditorContainer from "./YamlEditorContainer.svelte";

  export let fileName: string;

  let editor: YAMLEditor;
  let view: EditorView;

  const queryClient = useQueryClient();
  const saveFile = createRuntimeServicePutFileAndReconcile();

  $: file = createRuntimeServiceGetFile($runtime.instanceId, fileName, {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  let error;

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    error = undefined;
    const blob = e.detail.content;

    // Save File
    $saveFile.mutate({
      data: {
        instanceId: $runtime.instanceId,
        path: fileName,
        blob: blob,
      },
    });

    // Invalidate Get File
    queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey($runtime.instanceId, fileName)
    );

    // Get YAML syntax errors
    try {
      parse(blob);
    } catch (e) {
      error = e;
    }
  }
</script>

<YamlEditorContainer errorMessage={error}>
  <YAMLEditor
    bind:this={editor}
    bind:view
    content={$file?.data?.blob || ""}
    on:update={handleUpdate}
  />
</YamlEditorContainer>
