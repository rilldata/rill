<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import TableIcon from "../../../components/icons/TableIcon.svelte";
  import TableMenuItems from "./TableMenuItems.svelte";
  import TableSchema from "./TableSchema.svelte";
  import { useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector } from "../selectors";
  import { runtime } from "../../../runtime-client/runtime-store";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import {
    makeFullyQualifiedTableName,
    makeTablePreviewHref,
  } from "../connectors-utils";

  export let driver: string;
  export let connector: string;
  export let database: string; // The backend interprets an empty string as the default database
  export let databaseSchema: string; // The backend interprets an empty string as the default schema
  export let table: string;
  export let store: ConnectorExplorerStore;
  export let showGenerateMetricsAndDashboard: boolean = false;
  export let showGenerateModel: boolean = false;
  export let isOlapConnector: boolean = false;

  let contextMenuOpen = false;

  $: expandedStore = store.getItem(connector, database, databaseSchema, table);
  $: showSchema = $expandedStore;

  const { allowContextMenu, allowNavigateToTable, allowShowSchema } = store;

  $: ({ instanceId: runtimeInstanceId } = $runtime);
  $: isModelingSupportedForConnector = useIsModelingSupportedForConnector(
    runtimeInstanceId,
    connector,
  );
  $: isModelingSupported = $isModelingSupportedForConnector.data;

  $: fullyQualifiedTableName = makeFullyQualifiedTableName(
    driver,
    database,
    databaseSchema,
    table,
  );
  $: tableId = `${connector}-${fullyQualifiedTableName}`;

  // Generate preview href for any connector that supports preview routes
  $: href =
    makeTablePreviewHref(driver, connector, database, databaseSchema, table) ||
    undefined;

  $: open = href ? $page.url.pathname === href : false;

  // Allow navigation when a preview href is available
  $: element = allowNavigateToTable && href ? "a" : "button";
</script>

<li aria-label={tableId} class="table-entry group" class:open>
  <div
    class:pl-[58px]={database || !allowShowSchema}
    class="table-entry-header pl-10"
  >
    {#if allowShowSchema}
      <button
        on:click={() => {
          store.toggleItem(connector, database, databaseSchema, table);
        }}
      >
        <CaretDownIcon
          className="flex-none transform transition-transform text-gray-400 {!showSchema &&
            '-rotate-90'}"
          size="14px"
        />
      </button>
    {/if}

    <svelte:element
      this={element}
      class="clickable-text"
      {...allowNavigateToTable && href ? { href } : {}}
      role="menuitem"
      tabindex="0"
      on:click={() => {
        store.toggleItem(connector, database, databaseSchema, table);
      }}
    >
      <TableIcon size="14px" className="shrink-0 text-gray-400" />
      <span class="truncate">
        {table}
      </span>
    </svelte:element>

    {#if allowContextMenu && (showGenerateMetricsAndDashboard || isModelingSupported || showGenerateModel)}
      <DropdownMenu.Root bind:open={contextMenuOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <ContextButton
            id="more-actions-{tableId}"
            testId="more-actions-context-button"
            tooltipText="More actions"
            label="{tableId} actions menu trigger"
            builders={[builder]}
            suppressTooltip={contextMenuOpen}
          >
            <MoreHorizontal />
          </ContextButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          class="min-w-60"
          align="start"
          side="right"
          sideOffset={16}
        >
          <TableMenuItems
            {connector}
            {database}
            {databaseSchema}
            {table}
            {showGenerateMetricsAndDashboard}
            {showGenerateModel}
            {isModelingSupported}
            {isOlapConnector}
          />
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>

  {#if allowShowSchema && showSchema}
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

  .selected:hover {
    @apply bg-slate-200;
  }
</style>
