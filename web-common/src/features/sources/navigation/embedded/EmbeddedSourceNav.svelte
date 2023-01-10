<script lang="ts">
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { GroupedURIObject, groupURIs } from "../../group-uris";
  import { useEmbeddedSources } from "../../selectors";
  import EmbeddedSourceSet from "./EmbeddedSourceSet.svelte";

  $: sourceCatalogsQuery = useEmbeddedSources($runtimeStore?.instanceId);
  let embeddedSourceCatalogs: Array<V1CatalogEntry>;
  $: embeddedSourceCatalogs = $sourceCatalogsQuery?.data ?? [];

  $: chunksOfLinks = groupURIs(embeddedSourceCatalogs) as GroupedURIObject;
</script>

<div class="space-y-2">
  {#each Object.keys(chunksOfLinks) as domain, i (domain)}
    {@const domainSet = chunksOfLinks[domain]}
    {@const links = domainSet.uris}
    {@const connector = domainSet.connector}
    <EmbeddedSourceSet location={domain} {connector} sources={links} />
  {/each}
</div>
