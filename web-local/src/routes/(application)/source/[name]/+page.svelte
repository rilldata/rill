<script lang="ts">
  import { page } from "$app/stores";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
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
  $: filePath = getFileAPIPathFromNameAndType(sourceName, EntityType.Table);

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

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

  // Initialize the source store
  $: sourceStore = useSourceStore(filePath);
  $: if ($fileQuery.isSuccess && $fileQuery.data.blob) {
    sourceStore.set({ clientYAML: $fileQuery.data.blob });
  }
</script>

<svelte:head>
  <title>Rill Developer | {sourceName}</title>
</svelte:head>

<SourceWorkspace {filePath} />
