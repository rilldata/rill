<script lang="ts">
  import { page } from "$app/stores";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ModelWorkspace } from "@rilldata/web-common/features/models";
  import UnsavedSourceDialog from "@rilldata/web-common/features/sources/editor/UnsavedSourceDialog.svelte";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { goto, beforeNavigate } from "$app/navigation";
  import { hasUnsavedChanges } from "@rilldata/web-common/features/models/workspace/Editor.svelte";

  const { readOnly } = featureFlags;

  let interceptedUrl: string | null = null;

  $: modelName = $page.params.name;
  $: filePath = getFileAPIPathFromNameAndType(modelName, EntityType.Model);

  /** If ?focus, tell the page to focus the editor as soon as available */
  $: focusEditor = $page.url.searchParams.get("focus") === "";

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

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  beforeNavigate((e) => {
    if (!$hasUnsavedChanges || interceptedUrl) return;

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
  <title>Rill Developer | {modelName}</title>
</svelte:head>

{#if $fileQuery.data}
  <ModelWorkspace {filePath} focusEditorOnMount={focusEditor} />
{/if}

{#if interceptedUrl}
  <UnsavedSourceDialog
    context="model"
    on:confirm={handleConfirm}
    on:cancel={handleCancel}
  />
{/if}
