<script lang="ts">
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import type { EditorView } from "@codemirror/view";
  import { parse } from "yaml";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { saveFile } from "@rilldata/web-common/features/generic-yaml-editor/actions";
  import ErrorPane from "@rilldata/web-common/features/generic-yaml-editor/ErrorPane.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  const fileName = "rill.yaml";

  export let data;

  let editor: YAMLEditor;
  let view: EditorView;
  let error: Error | undefined;

  async function handleUpdate(e: CustomEvent<{ content: string }>) {
    const blob = e.detail.content;
    await saveFile(queryClient, fileName, blob);
    error = validateYAMLAndReturnError(blob);
  }

  function validateYAMLAndReturnError(blob: string): Error | undefined {
    try {
      parse(blob);
      return undefined;
    } catch (e) {
      return e;
    }
  }

  function cleanErrorMessage(message: string): string {
    return message?.replace("YAMLParseError: ", "");
  }
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    slot="header"
    titleInput={fileName}
    editable={false}
    showInspectorToggle={false}
  />
  <div slot="body" class="flex flex-col size-full bg-white">
    <YAMLEditor
      bind:this={editor}
      bind:view
      content={data.blob || ""}
      on:update={handleUpdate}
    />

    {#if error}
      <ErrorPane errorMessage={cleanErrorMessage(error.message)} />
    {/if}
  </div>
</WorkspaceContainer>
