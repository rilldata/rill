<script lang="ts">
  import { page } from "$app/stores";
  import SourceEmbedded from "@rilldata/web-common/components/icons/SourceEmbedded.svelte";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  import { slide } from "svelte/transition";
  import EmbeddedSourceMenuItems from "./EmbeddedSourceMenuItems.svelte";
  import EmbeddedSourceTooltip from "./EmbeddedSourceTooltip.svelte";

  export let embeddedSourceCatalog: V1CatalogEntry;
  $: uri = embeddedSourceCatalog.source.properties.path;

  $: cachedSourceName = embeddedSourceCatalog.name;
  $: connector = embeddedSourceCatalog.source.connector;
</script>

<NavigationEntry
  href={`/source/${cachedSourceName}`}
  open={$page.url.pathname === `/source/${cachedSourceName}`}
  name={uri}
  tooltipMaxWidth="300px"
  maxMenuWidth="300px"
>
  <SourceEmbedded slot="icon" />
  <svelte:fragment slot="more">
    <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
      <ColumnProfile indentLevel={1} objectName={uri} />
    </div>
  </svelte:fragment>

  <svelte:fragment slot="tooltip-content">
    <EmbeddedSourceTooltip sourceName={uri} {connector} />
  </svelte:fragment>

  <svelte:fragment slot="menu-items">
    <EmbeddedSourceMenuItems {cachedSourceName} {uri} {connector} />
  </svelte:fragment>
</NavigationEntry>
