<script lang="ts">
  import { page } from "$app/stores";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";

  $: fileQuery = createRuntimeServiceGetFile("default", $page.params.file);
  $: resourceKind = getResourceKindFromFile($page.params.file);

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
  // const putFile = createRuntimeServicePutFileAndReconcile();
</script>

<!-- on:write={(evt) => $putFile.mutate(evt.detail.blob)} -->
{#if isSource || isModel}
  <!-- use the workspace +page.svelte file -->
{:else if isDashboard}
  <!-- use the Metrics View +page.svelte file -->
{:else if isUnknown}
  <WorkspaceContainer>
    <FilesWorkspaceHeader filePath={$page.params.file} slot="header" />
    <Editor
      content={$fileQuery.data?.blob}
      on:write={(evt) => console.log(evt)}
      slot="body"
    />
  </WorkspaceContainer>
{/if}
