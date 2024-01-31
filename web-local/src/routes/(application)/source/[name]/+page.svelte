<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { SourceWorkspace } from "@rilldata/web-common/features/sources";
  import { useSourceStore } from "@rilldata/web-common/features/sources/sources-store";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;

  $: sourceName = $page.params.name;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
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
    },
  );

  // Initialize the source store
  $: sourceStore = useSourceStore(sourceName);
  $: if ($fileQuery.isSuccess && $fileQuery.data.blob) {
    sourceStore.set({ clientYAML: $fileQuery.data.blob });
  }
</script>

<svelte:head>
  <title>Rill Developer | {sourceName}</title>
</svelte:head>

<SourceWorkspace {sourceName} />
