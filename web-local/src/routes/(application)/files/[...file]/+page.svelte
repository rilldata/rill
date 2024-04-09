<script lang="ts">
  import { page } from "$app/stores";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: filePath = $page.params.file;
  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);
  $: resourceKind = getResourceKindFromFile(filePath);

  $: isSource = resourceKind === "source";
  $: isModel = resourceKind === "model";
  $: isDashboard = resourceKind === "dashboard";
  $: isUnknown = !isSource && !isModel && !isDashboard;

  function getResourceKindFromFile(filePath: string) {
    // remove file extension
    const pathWithoutExtension = filePath.replace(/\.[^/.]+$/, "");
    // get the last part of the path
    const pathParts = pathWithoutExtension.split("/");
    return pathParts[pathParts.length - 1];
  }
  // TODO: optimistically update the get file cache
  const putFile = createRuntimeServicePutFile();

  function handleFileUpdate(content: string) {
    if ($fileQuery.data?.blob === content) return;
    return $putFile.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        blob: content,
      },
      path: filePath,
    });
  }
</script>

<!-- on:write={(evt) => $putFile.mutate(evt.detail.blob)} -->
{#if isSource || isModel}
  <!-- use the workspace +page.svelte file -->
{:else if isDashboard}
  <!-- use the Metrics View +page.svelte file -->
{:else if isUnknown}
  <WorkspaceContainer>
    <FileWorkspaceHeader filePath={$page.params.file} slot="header" />
    <Editor
      content={$fileQuery.data?.blob ?? ""}
      on:write={({ detail: { content } }) => handleFileUpdate(content)}
      slot="body"
    />
  </WorkspaceContainer>
{/if}
