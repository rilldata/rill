<script lang="ts">
  import PlusIcon from "@rilldata/web-common/components/icons/PlusIcon.svelte";
  import ConnectorExplorer from "@rilldata/web-common/features/connectors/explorer/ConnectorExplorer.svelte";
  import { ConnectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import QueryCell from "./QueryCell.svelte";
  import QuerySchemaPanel from "./QuerySchemaPanel.svelte";
  import { makeSufficientlyQualifiedTableName } from "@rilldata/web-common/features/connectors/connectors-utils";
  import {
    createNotebook,
    type NotebookStore,
    type NotebookState,
  } from "./query-store";
  import { onDestroy } from "svelte";
  import { get, readable } from "svelte/store";

  const WORKSPACE_KEY = "__query_console__";

  export let projectId = "";

  const runtimeClient = useRuntimeClient();

  // Get default OLAP connector for new cells (sensitive: true required to include olapConnector)
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: olapConnector = $instanceQuery.data?.instance?.olapConnector ?? "";

  // Create notebook store once we have the default connector
  let notebook: NotebookStore | null = null;
  $: if (olapConnector && !notebook) {
    notebook = createNotebook(olapConnector, projectId);
  }

  onDestroy(() => notebook?.destroy());

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
    // onInsertTable: "+" button inserts SELECT * FROM table at cursor in focused cell
    (driver, _connector, database, schema, table) => {
      if (!notebook) return;
      const state = get(notebook);
      const focusedId = state.focusedCellId ?? state.cells[0]?.id;
      if (!focusedId) return;

      const tableRef = makeSufficientlyQualifiedTableName(
        driver,
        database,
        schema,
        table,
      );
      const sql = `SELECT * FROM ${tableRef}`;

      cellRefs[focusedId]?.insertAtCursor(sql);
    },
  );

  let sidebarWidth = 260;

  // Fallback stores for when notebook is null (Svelte can't auto-subscribe nullable stores)
  const EMPTY_NOTEBOOK = readable<NotebookState>({
    cells: [],
    focusedCellId: null,
  });
  const NULL_READABLE = readable<null>(null);
  const ZERO_READABLE = readable<number>(0);

  // Always-valid store references for $-prefix subscriptions
  $: nb = notebook ?? EMPTY_NOTEBOOK;
  $: cells = $nb.cells;

  // Clean up stale refs when cells change
  $: {
    const cellIds = new Set(cells.map((c) => c.id));
    for (const id of Object.keys(cellRefs)) {
      if (!cellIds.has(id)) delete cellRefs[id];
    }
  }

  // Derived stores for the focused cell (forwarded to inspector)
  $: focusedSchemaStore = notebook?.focusedSchema ?? NULL_READABLE;
  $: focusedRowCountStore = notebook?.focusedRowCount ?? ZERO_READABLE;
  $: focusedTimeMsStore = notebook?.focusedExecutionTimeMs ?? NULL_READABLE;

  function handleAddCell() {
    notebook?.addCell(olapConnector);
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

{#if notebook}
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
        {#each cells as cell (cell.id)}
          <QueryCell
            bind:this={cellRefs[cell.id]}
            cellId={cell.id}
            {notebook}
            cellCount={cells.length}
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
      schema={$focusedSchemaStore}
      rowCount={$focusedRowCountStore}
      executionTimeMs={$focusedTimeMsStore}
      {selectedTable}
    />
  </div>
{/if}

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
