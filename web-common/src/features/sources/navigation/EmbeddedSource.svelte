<script lang="ts">
  import { page } from "$app/stores";
  import SourceEmbedded from "@rilldata/web-common/components/icons/SourceEmbedded.svelte";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  import { slide } from "svelte/transition";
  import EmbeddedSourceTooltip from "./EmbeddedSourceTooltip.svelte";

  export let embeddedSourceCatalog: V1CatalogEntry;
  $: sourceName = embeddedSourceCatalog.source.properties.path;
  $: cachedTableName = embeddedSourceCatalog.name;
  $: connector = embeddedSourceCatalog.source.connector;
  $: embeds = embeddedSourceCatalog.embeds;
</script>

<NavigationEntry
  href={`/source/${cachedTableName}`}
  open={$page.url.pathname === `/source/${cachedTableName}`}
  name={sourceName}
  tooltipMaxWidth="300px"
>
  <SourceEmbedded slot="icon" />
  <svelte:fragment slot="more">
    <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
      <ColumnProfile indentLevel={1} objectName={sourceName} />
    </div>
  </svelte:fragment>

  <svelte:fragment slot="tooltip-content">
    <EmbeddedSourceTooltip {sourceName} {connector} {embeds} />
  </svelte:fragment>

  <svelte:fragment slot="menu-items" let:toggleMenu>
    <!-- <SourceMenuItems
              {sourceName}
              {toggleMenu}
              on:rename-asset={() => {
                openRenameTableModal(sourceName);
              }}
            /> -->
  </svelte:fragment>
</NavigationEntry>
