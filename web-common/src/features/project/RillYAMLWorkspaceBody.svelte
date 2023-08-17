<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "../../components/editor/YAMLEditor.svelte";
  import { createRuntimeServiceGetFile } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";

  let editor: YAMLEditor;
  let view: EditorView;

  $: file = createRuntimeServiceGetFile($runtime.instanceId, "rill.yaml", {
    query: {
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  function handleUpdate(e: CustomEvent<{ content: string }>) {
    console.log(e.detail.content);
    // // Update the client-side store
    // sourceStore.set({ clientYAML: e.detail.content });

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
