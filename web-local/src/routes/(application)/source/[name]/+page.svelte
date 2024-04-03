<script lang="ts">
  import UnsavedSourceDialog from "@rilldata/web-common/features/sources/editor/UnsavedSourceDialog.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import SourceInspector from "@rilldata/web-common/features/sources/inspector/SourceInspector.svelte";
  import SourceWorkspaceHeader from "@rilldata/web-common/features/sources/workspace/SourceWorkspaceHeader.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import ErrorPane from "@rilldata/web-common/features/generic-yaml-editor/ErrorPane.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { saveAndRefresh } from "@rilldata/web-common/features/sources/saveAndRefresh";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;

  let latest: string;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: sourceName = $page.params.name;
  $: filePath = getFileAPIPathFromNameAndType(sourceName, EntityType.Table);

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      onError: (err) => {
        if (err.response?.status && err.response?.data?.message) {
          throw error(err.response.status, err.response.data.message);
        } else {
          console.error(err);
          throw error(500, err.message);
        }
      },
    },
  });

  $: isSourceUnsaved = latest !== $fileQuery.data?.blob;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  $: sourceQuery = fileArtifact.getResource(queryClient, $runtime.instanceId);

  function revert() {
    latest = $fileQuery.data?.blob ?? "";
  }

  async function save() {
    overlay.set({ title: `Importing ${filePath}` });
    await saveAndRefresh(filePath, latest);
    checkSourceImported(queryClient, filePath);
    overlay.set(null);
  }

  let interceptedUrl: string | null = null;

  beforeNavigate((e) => {
    if (!isSourceUnsaved || interceptedUrl) return;

    e.cancel();

    if (e.to) {
      interceptedUrl = e.to.url.href;
    }
  });

  async function handleConfirm() {
    if (interceptedUrl) {
      await goto(interceptedUrl);
    }

    interceptedUrl = null;
  }

  function handleCancel() {
    interceptedUrl = null;
  }
</script>

<svelte:head>
  <title>Rill Developer | {sourceName}</title>
</svelte:head>

<WorkspaceContainer>
  <SourceWorkspaceHeader
    slot="header"
    {filePath}
    {sourceName}
    {isSourceUnsaved}
    on:revert={revert}
    on:save={save}
  />
  <div
    class="editor-pane h-full overflow-hidden w-full flex flex-col"
    slot="body"
  >
    <WorkspaceEditorContainer>
      <SourceEditor
        {isSourceUnsaved}
        {filePath}
        yaml={$fileQuery.data?.blob ?? ""}
        bind:latest
      />
    </WorkspaceEditorContainer>

    <WorkspaceTableContainer fade={isSourceUnsaved}>
      {#if !$allErrors?.length}
        <ConnectedPreviewTable
          objectName={$sourceQuery?.data?.source?.state?.table}
          loading={resourceIsLoading($sourceQuery?.data)}
        />
      {:else if $allErrors[0].message}
        <ErrorPane errorMessage={$allErrors[0].message} />
      {/if}
    </WorkspaceTableContainer>
  </div>
  <SourceInspector {filePath} slot="inspector" {isSourceUnsaved} />
</WorkspaceContainer>

{#if interceptedUrl}
  <UnsavedSourceDialog
    context="model"
    on:confirm={handleConfirm}
    on:cancel={handleCancel}
  />
{/if}
