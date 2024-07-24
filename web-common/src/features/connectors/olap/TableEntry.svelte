<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import TableIcon from "../../../components/icons/TableIcon.svelte";
  import TableMenuItems from "./TableMenuItems.svelte";
  import TableSchema from "./TableSchema.svelte";
  import UnsupportedTypesIndicator from "./UnsupportedTypesIndicator.svelte";
  import {
    makeFullyQualifiedTableName,
    makeTablePreviewHref,
  } from "./olap-config";

  export let instanceId: string;
  export let driver: string;
  export let connector: string;
  export let database: string; // The backend interprets an empty string as the default database
  export let databaseSchema: string; // The backend interprets an empty string as the default schema
  export let table: string;
  export let hasUnsupportedDataTypes: boolean;

  let contextMenuOpen = false;
  let showSchema = false;

  $: fullyQualifiedTableName = makeFullyQualifiedTableName(
    driver,
    database,
    databaseSchema,
    table,
  );
  $: tableId = `${connector}-${fullyQualifiedTableName}`;
  $: href = makeTablePreviewHref(
    driver,
    connector,
    database,
    databaseSchema,
    table,
  );
  $: open = $page.url.pathname === href;
</script>

<li aria-label={tableId} class="table-entry group" class:open>
  <div class="table-entry-header {database ? 'pl-[58px]' : 'pl-[40px]'}">
    <button on:click={() => (showSchema = !showSchema)}>
      <CaretDownIcon
        className="transform transition-transform text-gray-400 {showSchema
          ? 'rotate-0'
          : '-rotate-90'}"
        size="14px"
      />
    </button>
    <a class="clickable-text" {href}>
      <TableIcon size="14px" className="shrink-0 text-gray-400" />
      <span class="truncate">
        {table}
      </span>
    </a>
    {#if hasUnsupportedDataTypes}
      <UnsupportedTypesIndicator
        {instanceId}
        {connector}
        {database}
        {databaseSchema}
        {table}
      />
    {/if}
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          id="more-actions-{tableId}"
          tooltipText="More actions"
          label="{tableId} actions menu trigger"
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
  .table-entry {
    @apply w-full justify-between;
    @apply flex flex-col;
  }

  .table-entry-header {
    @apply h-6 pr-2; /* left-padding is set dynamically above */
    @apply flex justify-between items-center gap-x-1;
  }

  .table-entry-header:hover {
    @apply bg-slate-100;
  }

  .open {
    @apply bg-slate-100;
  }

  .clickable-text {
    @apply flex grow items-center gap-x-1;
    @apply text-gray-900 truncate;
  }
</style>
