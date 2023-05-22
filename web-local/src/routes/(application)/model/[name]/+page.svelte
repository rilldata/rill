<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ModelWorkspace } from "@rilldata/web-common/features/models";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  $: modelName = $page.params.name;

  /** If ?focus, tell the page to focus the editor as soon as available */
  $: focusEditor = $page.url.searchParams.get("focus") === "";

  onMount(() => {
    if ($featureFlags.readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(modelName, EntityType.Model),
    {
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
    }
  );
</script>

<svelte:head>
  <title>Rill Developer | {modelName}</title>
</svelte:head>

{#if $fileQuery.data}
  <ModelWorkspace {modelName} focusEditorOnMount={focusEditor} />
{/if}
