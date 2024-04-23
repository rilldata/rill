<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import Tooltip from "../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../components/tooltip/TooltipContent.svelte";
  import TableMenuItems from "./TableMenuItems.svelte";
  import TableSchema from "./TableSchema.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import {
    makeFullyQualifiedTableName,
    makeTablePreviewHref,
  } from "./olap-config";

  export let connectorInstanceId: string;
  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;
  export let hasUnsupportedDataTypes: boolean;

  let contextMenuOpen = false;
  let showSchema = false;

  $: fullyQualifiedTableName = makeFullyQualifiedTableName(
    connector,
    database,
    databaseSchema,
    table,
  );
  $: href = makeTablePreviewHref(connector, database, databaseSchema, table);
  $: open = $page.url.pathname === href;
</script>

<li aria-label={fullyQualifiedTableName} class="entry group">
  <div
    class="header-wrapper flex justify-between items-center gap-x-2 pl-2 pr-2"
    class:open
  >
    <TableIcon size="14px" className="shrink-0 text-gray-400" />
    <Tooltip alignment="start" distance={4}>
      <button
        class="clickable-text"
        on:click={() => (showSchema = !showSchema)}
      >
        <span class="truncate">
          {fullyQualifiedTableName}
        </span>
        {#if hasUnsupportedDataTypes}
          <UnsupportedTypesIndicator
            instanceId={connectorInstanceId}
            {connector}
            {database}
            {databaseSchema}
            {table}
          />
        {/if}
      </button>
      <TooltipContent slot="tooltip-content">
        {showSchema ? "Hide schema" : "Show schema"}
      </TooltipContent>
    </Tooltip>
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          id="more-actions-{name}"
          tooltipText="More actions"
          label="{name} actions menu trigger"
          builders={[builder]}
          suppressTooltip={contextMenuOpen}
        >
          <MoreHorizontal />
        </ContextButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        class="border-none bg-gray-800 text-white min-w-60"
        align="start"
        side="right"
        sideOffset={16}
      >
        <TableMenuItems {connector} {database} {databaseSchema} {table} />
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>
  {#if showSchema}
    <TableSchema {connector} {database} {databaseSchema} {table} />
  {/if}
</li>

<style lang="postcss">
  .entry {
    @apply w-full justify-between;
    @apply flex flex-col;
  }

  .entry:focus {
    @apply outline-none;
  }

  /* .entry:hover:not(.open) {
    @apply text-gray-900;
  } */

  .entry:focus,
  .open {
    @apply bg-slate-100;
  }

  .clickable-text {
    @apply select-none cursor-pointer;
    @apply w-full flex items-center gap-x-2 truncate;
    @apply text-gray-900;
  }
  .clickable-text:hover {
    @apply text-gray-900;
  }

  .header-wrapper {
    @apply sticky top-0;
    @apply z-10 bg-white;
  }
</style>
