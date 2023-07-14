<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { SourceWorkspace } from "@rilldata/web-common/features/sources";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  $: sourceName = $page.params.name;

  onMount(() => {
    if ($featureFlags.readOnly) {
      throw error(404, "Page not found");
    }
  });

  // try to get the catalog entry.
  $: catalogQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    sourceName,
    {
      query: {
        onError: () => {
          // no-op. we'll try to get the file below.
        },
      },
    }
  );
  $: embedded = $catalogQuery.data?.entry?.embedded;
  $: path = $catalogQuery.data?.entry?.source.properties.path;

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
    }
  );
</script>

<svelte:head>
  <title>Rill Developer | {embedded ? path : sourceName}</title>
</svelte:head>

{#if $fileQuery.data && $catalogQuery.data}
  <SourceWorkspace {sourceName} {embedded} {path} />
{/if}
