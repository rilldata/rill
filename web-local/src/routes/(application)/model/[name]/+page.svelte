<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { ModelWorkspace } from "@rilldata/web-common/features/models";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { featureFlags } from "../../../../lib/application-state-stores/application-store";

  const modelName: string = $page.params.name;

  /** If ?focus, tell the page to focus the editor as soon as available */
  const focusEditor = $page.url.searchParams.get("focus") === "";

  onMount(() => {
    if ($featureFlags.readOnly) {
      throw error(404, "Page not found");
    }
  });

  const fileQuery = useRuntimeServiceGetFile(
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
