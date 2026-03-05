<script lang="ts">
  import PlusIcon from "@rilldata/web-common/components/icons/PlusIcon.svelte";
  import ConnectorExplorer from "@rilldata/web-common/features/connectors/explorer/ConnectorExplorer.svelte";
  import { ConnectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import QueryCell from "./QueryCell.svelte";
  import QuerySchemaPanel from "./QuerySchemaPanel.svelte";
  import { makeSufficientlyQualifiedTableName } from "@rilldata/web-common/features/connectors/connectors-utils";
  import { createNotebook } from "./query-store";

  const WORKSPACE_KEY = "__query_console__";

  $: ({ instanceId } = $runtime);

  // Get default OLAP connector for new cells
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instanceQuery.data?.instance?.olapConnector ?? "";

  // Create notebook store once we have the default connector
  let notebook = createNotebook("");
  let initialized = false;
  $: if (olapConnector && !initialized) {
    notebook = createNotebook(olapConnector);
    initialized = true;
  }

  // Track table selected from data explorer (for ColumnProfile in right panel)
  let selectedTable: {
    connector: string;
    database: string;
    databaseSchema: string;
    objectName: string;
  } | null = null;

  // Refs to cell editors for programmatic content setting
  let cellRefs: Record<string, QueryCell> = {};

  // Data explorer sidebar
  const explorerStore = new ConnectorExplorerStore(
    {
      allowNavigateToTable: false,
      allowContextMenu: false,
      allowShowSchema: true,
      allowSelectTable: false,
    },
    // onToggleItem: show ColumnProfile when a table is expanded
    (connector, database, schema, table) => {
      if (!table) {
        selectedTable = null;
        return;
      }
      selectedTable = {
        connector,
        database: database ?? "",
        databaseSchema: schema ?? "",
        objectName: table,
      };
    },
    // onInsertTable: "+" button populates the focused cell
    (driver, connector, database, schema, table) => {
      const tableRef = makeSufficientlyQualifiedTableName(
        driver,
        database,
        schema,
        table,
      );
      const sql = `SELECT * FROM ${tableRef}`;

      const focusedId = $notebook.focusedCellId ?? $notebook.cells[0]?.id;
      if (focusedId) {
        notebook.setCellConnector(focusedId, connector);
        notebook.setCellSql(focusedId, sql);
        cellRefs[focusedId]?.setEditorContent(sql);
      }
    },
  );

  let sidebarWidth = 260;

  // Derived stores for the focused cell (forwarded to inspector)
  $: focusedSchema = notebook.focusedSchema;
  $: focusedRowCount = notebook.focusedRowCount;
  $: focusedExecutionTimeMs = notebook.focusedExecutionTimeMs;

  function handleAddCell() {
    notebook.addCell(olapConnector);
  }

  function handleCellRun() {
    // Clear table selection when a query is run (show query results instead)
    selectedTable = null;
  }

  function handleCellFocus() {
    // Clear table selection so the right panel shows query result schema
    selectedTable = null;
  }
</script>

<div class="query-workspace">
  <!-- Left Sidebar: Data Explorer -->
  <aside class="data-explorer" style:width="{sidebarWidth}px">
    <div class="sidebar-header">
      <h3
        class="text-xs font-semibold text-fg-secondary uppercase tracking-wide"
      >
        Data Explorer
      </h3>
    </div>
    <div class="sidebar-content">
      <ConnectorExplorer store={explorerStore} />
    </div>
  </aside>

  <Resizer
    absolute={false}
    direction="EW"
    side="right"
    min={200}
    max={440}
    bind:dimension={sidebarWidth}
  />

  <!-- Center: Notebook cells -->
  <div class="notebook-area">
    <div class="cells-container">
      {#each $notebook.cells as cell (cell.id)}
        <QueryCell
          bind:this={cellRefs[cell.id]}
          cellId={cell.id}
          {notebook}
          {instanceId}
          cellCount={$notebook.cells.length}
          on:focus={handleCellFocus}
          on:run={handleCellRun}
        />
      {/each}

      <button class="add-cell-button" on:click={handleAddCell}>
        <PlusIcon size="14px" />
        Add Cell
      </button>
    </div>
  </div>

  <!-- Right Sidebar: Schema Inspector -->
  <QuerySchemaPanel
    filePath={WORKSPACE_KEY}
    schema={$focusedSchema}
    rowCount={$focusedRowCount}
    executionTimeMs={$focusedExecutionTimeMs}
    {selectedTable}
  />
</div>

<style lang="postcss">
  .query-workspace {
    @apply flex size-full overflow-hidden bg-gray-100/80;
  }

  .data-explorer {
    @apply flex-none flex flex-col overflow-hidden;
    @apply border-r bg-surface-background;
  }

  .sidebar-header {
    @apply px-3 py-2 border-b;
  }

  .sidebar-content {
    @apply overflow-y-auto flex-1;
  }

  .notebook-area {
    @apply flex-1 overflow-hidden flex flex-col min-w-0;
  }

  .cells-container {
    @apply flex flex-col gap-y-3 p-4 overflow-y-auto h-full;
  }

  .add-cell-button {
    @apply flex items-center gap-x-1.5 justify-center;
    @apply w-full py-2 rounded border border-dashed;
    @apply text-xs text-fg-secondary;
  }

  .add-cell-button:hover {
    @apply bg-surface-subtle text-fg-primary border-solid;
  }
</style>
