<script lang="ts">
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import SourceIcon from "@rilldata/web-common/components/icons/Source.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let embeddedSourceCatalog: V1CatalogEntry;
</script>

<Tooltip alignment="start" distance={16} location="right">
  <div
    class="navigation-entry-title grid gap-x-1 items-center pl-4 pr-3 hover:bg-gray-100"
    on:command-click
    style:grid-template-columns="max-content auto max-content"
    style:height="24px"
  >
    <div class="embedded-source-icon">
      <SourceIcon />
    </div>
    <a class="ui-copy  text-ellipsis overflow-hidden whitespace-nowrap">
      {embeddedSourceCatalog.source.properties.path}
    </a>
  </div>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">Embedded source for</svelte:fragment>
    </TooltipTitle>
    {#each embeddedSourceCatalog.embeds as embeds (embeds)}
      {embeds}<br />
    {/each}
  </TooltipContent>
</Tooltip>

<style>
  .embedded-source-icon {
    margin: 0px 6px 0px 4px;
  }
</style>
