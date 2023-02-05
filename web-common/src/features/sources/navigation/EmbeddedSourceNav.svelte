<script lang="ts">
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import EmbeddedSourceNavigationEntry from "../embedded/EmbeddedSourceNavigationEntry.svelte";
  import { useEmbeddedSources } from "../selectors";

  $: sourceCatalogsQuery = useEmbeddedSources($runtimeStore?.instanceId);
  let embeddedSourceCatalogs: Array<V1CatalogEntry>;
  $: embeddedSourceCatalogs = $sourceCatalogsQuery?.data ?? [];
</script>

<div class="space-y-2">
  {#each embeddedSourceCatalogs as source, i (source.name)}
    <EmbeddedSourceNavigationEntry
      connector={source?.source?.connector}
      path={source.path}
      cachedSourceName={source.name}
    />
  {/each}
</div>
