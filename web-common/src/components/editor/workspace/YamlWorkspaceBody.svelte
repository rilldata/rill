<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { parse } from "yaml";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
    getRuntimeServiceGetFileQueryKey,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import YAMLEditor from "../YAMLEditor.svelte";
  import YamlEditorContainer from "./YamlEditorContainer.svelte";

  export let fileName: string;

  let editor: YAMLEditor;
  let view: EditorView;
  let errorMessage: string;

  const queryClient = useQueryClient();
  const saveFile = createRuntimeServicePutFile();

  $: file = createRuntimeServiceGetFile($runtime.instanceId, fileName, {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  async function handleUpdate(e: CustomEvent<{ content: string }>) {
    const blob = e.detail.content;

    // Save File
    await $saveFile.mutateAsync({
      instanceId: $runtime.instanceId,
      path: fileName,
      data: {
        blob: blob,
      },
    });

    // Invalidate `GetFile` query
    queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey($runtime.instanceId, fileName)
    );

    // Check for YAML syntax error
    try {
      parse(blob);

      // No error
      errorMessage = undefined;
    } catch (e) {
      // Error
      errorMessage = e.message;
    }
  }

  function cleanErrorMessage(message: string): string {
    return message?.replace("YAMLParseError: ", "");
  }
</script>

<YamlEditorContainer errorMessage={cleanErrorMessage(errorMessage)}>
  <YAMLEditor
    bind:this={editor}
    bind:view
    content={$file?.data?.blob || ""}
    on:update={handleUpdate}
  />
</YamlEditorContainer>
