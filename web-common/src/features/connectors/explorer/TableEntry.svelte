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
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import {
    makeFullyQualifiedTableName,
    makeTablePreviewHref,
  } from "../connectors-utils";

  let {
    driver,
    connector,
    database,
    databaseSchema,
    table,
    store,
    showGenerateMetricsAndDashboard = false,
    showGenerateModel = false,
    isOlapConnector = false,
  }: {
    driver: string;
    connector: string;
    database: string; // The backend interprets an empty string as the default database
    databaseSchema: string; // The backend interprets an empty string as the default schema
    table: string;
    store: ConnectorExplorerStore;
    showGenerateMetricsAndDashboard?: boolean;
    showGenerateModel?: boolean;
    isOlapConnector?: boolean;
  } = $props();

  let contextMenuOpen = $state(false);

  const client = useRuntimeClient();

  let expandedStore = $derived(
    store.getItem(connector, database, databaseSchema, table),
  );
  let showSchema = $derived($expandedStore);

  let selectedTableStore = $derived(store.selectedTableStore);
  let allowContextMenu = $derived(store.allowContextMenu);
  let allowNavigateToTable = $derived(store.allowNavigateToTable);
  let allowShowSchema = $derived(store.allowShowSchema);
  let onInsertTable = $derived(store.onInsertTable);
  let selectedTableState = $derived($selectedTableStore);
  let selectedConnector = $derived(selectedTableState.connector);
  let selectedDatabase = $derived(selectedTableState.database);
  let selectedSchema = $derived(selectedTableState.schema);
  let selectedTable = $derived(selectedTableState.table);
  let isSelected = $derived(
    selectedConnector === connector &&
      selectedDatabase === database &&
      selectedSchema === databaseSchema &&
      selectedTable === table,
  );

  let isModelingSupportedForConnector = $derived(
    useIsModelingSupportedForConnector(client, connector),
  );
  let isModelingSupported = $derived($isModelingSupportedForConnector.data);

  let fullyQualifiedTableName = $derived(
    makeFullyQualifiedTableName(driver, database, databaseSchema, table),
  );
  let tableId = $derived(`${connector}-${fullyQualifiedTableName}`);

  // Generate preview href for any connector that supports preview routes
  let href = $derived(
    makeTablePreviewHref(driver, connector, database, databaseSchema, table) ||
      undefined,
  );

  let open = $derived(
    isSelected || (href ? $page.url.pathname === href : false),
  );

  // Allow navigation when a preview href is available
  let element = $derived(allowNavigateToTable && href ? "a" : "button");
</script>

<li aria-label={tableId} class="table-entry group" class:open>
  <div
    class:pl-[58px]={database || !allowShowSchema}
    class="table-entry-header pl-10"
  >
    {#if allowShowSchema}
      <button
        type="button"
        onclick={() => {
          store.toggleItem(connector, database, databaseSchema, table);
        }}
      >
        <CaretDownIcon
          className="flex-none transform transition-transform text-fg-secondary {!showSchema &&
            '-rotate-90'}"
          size="14px"
        />
      </button>
    {/if}

    <svelte:element
      this={element}
      class="clickable-text"
      {...allowNavigateToTable && href ? { href } : { type: "button" }}
      role="menuitem"
      tabindex="0"
      onclick={() => {
        store.toggleItem(connector, database, databaseSchema, table);
      }}
    >
      <TableIcon size="14px" className="shrink-0 text-fg-secondary" />
      <span class="truncate">
        {table}
      </span>
    </svelte:element>

    {#if onInsertTable}
      <button
        class="insert-button"
        aria-label="Insert {table} into query"
        title="Insert into query"
        onclick={(e) => {
          e.stopPropagation();
          onInsertTable(driver, connector, database, databaseSchema, table);
        }}
      >
        +
      </button>
    {/if}

    {#if allowContextMenu && (showGenerateMetricsAndDashboard || isModelingSupported || showGenerateModel)}
      <DropdownMenu.Root bind:open={contextMenuOpen}>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            <ContextButton
              {...props}
              data-testid="more-actions-context-button"
              tooltipText="More actions"
              label="{tableId} actions menu trigger"
              suppressTooltip={contextMenuOpen}
            >
              <MoreHorizontal />
            </ContextButton>
          {/snippet}
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
    @apply bg-surface-hover;
  }

  .open {
    @apply bg-gray-100;
  }

  .clickable-text {
    @apply flex grow items-center gap-x-1;
    @apply text-fg-primary truncate;
  }

  .selected:hover {
    @apply bg-gray-200;
  }

  .insert-button {
    @apply hidden flex-none items-center justify-center;
    @apply w-5 h-5 rounded text-xs font-semibold;
    @apply text-fg-secondary bg-transparent;
  }

  .table-entry-header:hover .insert-button,
  .open .insert-button {
    @apply flex;
  }

  .insert-button:hover {
    @apply text-fg-primary bg-gray-200;
  }
</style>
