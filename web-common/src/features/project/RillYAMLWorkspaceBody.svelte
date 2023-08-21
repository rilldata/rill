<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { useQueryClient } from "@tanstack/svelte-query";
  import YAMLEditor from "../../components/editor/YAMLEditor.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "../../runtime-client";
  import { invalidateRillYAML } from "../../runtime-client/invalidation";
  import { runtime } from "../../runtime-client/runtime-store";

  let editor: YAMLEditor;
  let view: EditorView;

  const queryClient = useQueryClient();
  const putFile = createRuntimeServicePutFile();

  $: file = createRuntimeServiceGetFile($runtime.instanceId, "rill.yaml", {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    const blob = e.detail.content;

    // Put File
    $putFile.mutate({
      instanceId: $runtime.instanceId,
      path: "rill.yaml",
      data: {
        blob: blob,
      },
    });

    // Invalidate Get File
    invalidateRillYAML(queryClient, $runtime.instanceId);

    // // Clear line errors (it's confusing when they're outdated)
    // setLineStatuses([], view);
  }
</script>

<div class="h-full bg-white overflow-y-auto">
  <YAMLEditor
    bind:this={editor}
    bind:view
    content={$file?.data?.blob || ""}
    on:update={handleUpdate}
  />
</div>
