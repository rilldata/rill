<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { parse } from "yaml";
  import YAMLEditor from "../../components/editor/YAMLEditor.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
    getRuntimeServiceGetFileQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import ErrorPane from "./ErrorPane.svelte";

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

<div
  class="flex flex-col w-full h-full content-stretch"
  style:height={"calc(100vh - var(--header-height))"}
>
  <div class="grow bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={errorMessage}
      class:border-red-500={errorMessage}
    >
      <YAMLEditor
        bind:this={editor}
        bind:view
        content={$file?.data?.blob || ""}
        on:update={handleUpdate}
      />
    </div>
  </div>
  {#if errorMessage}
    <ErrorPane errorMessage={cleanErrorMessage(errorMessage)} />
  {/if}
</div>
