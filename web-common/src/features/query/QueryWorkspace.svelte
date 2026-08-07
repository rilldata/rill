<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import ConnectorExplorer from "@rilldata/web-common/features/connectors/explorer/ConnectorExplorer.svelte";
  import { ConnectorExplorerStore } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { readable } from "svelte/store";
  import QueryEditor from "./QueryEditor.svelte";
  import QueryResultsTable from "./QueryResultsTable.svelte";
  import {
    createSqlWorksheet,
    SQL_CONSOLE_ROW_LIMIT,
    type SqlWorksheet,
    type SqlWorksheetState,
  } from "./query-store";
  import { formatExecutionTime, type SqlExecution } from "./query-utils";

  let { cacheKey, connector }: { cacheKey: string; connector: string } =
    $props();

  const runtimeClient = useRuntimeClient();
  const explorerStore = new ConnectorExplorerStore({
    allowNavigateToTable: false,
    allowContextMenu: false,
    allowShowSchema: true,
    allowSelectTable: false,
    localStorage: false,
  });
  const emptyWorksheet = readable<SqlWorksheetState>({
    sql: "",
    isExecuting: false,
    hasExecuted: false,
    result: null,
    error: null,
    executionTimeMs: null,
    lastExecution: null,
  });

  let worksheet: SqlWorksheet | null = $state.raw(null);
  let activeWorksheet = $derived(worksheet ?? emptyWorksheet);
  let {
    sql: worksheetSql,
    isExecuting,
    hasExecuted,
    result,
    error: worksheetError,
    executionTimeMs,
    lastExecution,
  } = $derived($activeWorksheet);
  let editor: { runCurrent(): void } | undefined = $state.raw();

  let sidebarWidth = $state(260);
  let resultsHeight = $state(280);
  let showTables = $state(true);

  $effect(() => {
    const currentWorksheet = createSqlWorksheet(connector, cacheKey);
    worksheet = currentWorksheet;

    return () => {
      currentWorksheet.destroy();
      worksheet = null;
    };
  });

  function execute(execution: SqlExecution) {
    if (!worksheet || isExecuting) return;
    void worksheet.execute(runtimeClient, execution);
  }

  function runCurrentStatement() {
    editor?.runCurrent();
  }
</script>

<div class="query-workspace">
  {#if worksheet}
    <div class="workspace-body">
      {#if showTables}
        <aside class="table-browser" style:width="{sidebarWidth}px">
          <header class="panel-header">
            <div>
              <h2>Tables</h2>
              <p>{connector}</p>
            </div>
            <Button
              type="toolbar"
              square
              small
              label="Hide tables"
              onClick={() => (showTables = false)}
            >
              <HideSidebar size="15px" open />
            </Button>
          </header>
          <div class="table-list">
            <ConnectorExplorer
              store={explorerStore}
              olapOnly={true}
              onlyConnector={connector}
              defaultExpanded={connector}
            />
          </div>
        </aside>

        <Resizer
          absolute={false}
          direction="EW"
          side="right"
          min={200}
          max={420}
          bind:dimension={sidebarWidth}
        />
      {/if}

      <main class="console-main">
        <header class="console-header">
          <div class="flex items-center gap-x-2">
            {#if !showTables}
              <Button
                type="toolbar"
                square
                small
                label="Show tables"
                onClick={() => (showTables = true)}
              >
                <HideSidebar size="15px" open={false} />
              </Button>
            {/if}
            <h1>SQL Console</h1>
            <span class="read-only-badge">Read-only</span>
          </div>

          <div class="flex items-center gap-x-2">
            {#if isExecuting}
              <span class="execution-status">
                <Spinner size="12px" status={EntityStatus.Running} />
                Running
              </span>
              <Button
                type="secondary-destructive"
                small
                onClick={() => worksheet?.cancel()}
              >
                Stop
              </Button>
            {:else}
              <span class="shortcut-hint">⌘/Ctrl + Enter</span>
              <Button
                type="primary"
                small
                disabled={!worksheetSql.trim()}
                onClick={runCurrentStatement}
              >
                Run
              </Button>
            {/if}
          </div>
        </header>

        <section class="worksheet-pane" aria-label="SQL worksheet">
          <div class="editor-pane">
            <QueryEditor
              bind:this={editor}
              initialValue={worksheetSql}
              disabled={isExecuting}
              onchange={(sql: string) => worksheet?.setSql(sql)}
              onrun={execute}
            />
          </div>

          {#if worksheetError}
            <div class="worksheet-error" role="alert">
              <CancelCircle className="text-destructive flex-none" />
              <span>{worksheetError}</span>
            </div>
          {/if}
        </section>

        <div class="results-resizer">
          <Resizer
            absolute={false}
            direction="NS"
            side="top"
            min={160}
            max={600}
            basis={280}
            hang={false}
            bind:dimension={resultsHeight}
          />
        </div>

        <section
          class="results-pane"
          style:height="{resultsHeight}px"
          aria-label="Query results"
        >
          <header class="results-header">
            <div class="results-title">
              <h2>Results</h2>
              {#if lastExecution}
                <span>
                  {lastExecution.startLine === lastExecution.endLine
                    ? `Line ${lastExecution.startLine}`
                    : `Lines ${lastExecution.startLine}–${lastExecution.endLine}`}
                </span>
              {/if}
            </div>

            <div class="results-meta">
              {#if isExecuting}
                <span class="flex items-center gap-x-1">
                  <Spinner size="12px" status={EntityStatus.Running} />
                  Running
                </span>
              {:else if result}
                <span>
                  {#if (result.data?.length ?? 0) === SQL_CONSOLE_ROW_LIMIT}
                    Showing first {formatInteger(SQL_CONSOLE_ROW_LIMIT)} rows
                  {:else}
                    {formatInteger(result.data?.length ?? 0)}
                    {(result.data?.length ?? 0) === 1 ? "row" : "rows"}
                  {/if}
                  {#if executionTimeMs !== null}
                    · {formatExecutionTime(executionTimeMs)}
                  {/if}
                </span>
              {/if}
            </div>
          </header>

          <div class="results-table">
            <QueryResultsTable {result} {hasExecuted} error={worksheetError} />
          </div>
        </section>
      </main>
    </div>
  {/if}
</div>

<style lang="postcss">
  .query-workspace {
    @apply relative size-full min-h-0 overflow-hidden bg-surface-subtle;
  }

  .workspace-body {
    @apply flex size-full min-h-0 overflow-hidden;
  }

  .table-browser {
    @apply flex min-w-0 flex-none flex-col overflow-hidden border-r bg-surface-background;
  }

  .panel-header,
  .console-header,
  .results-header {
    @apply flex h-10 flex-none items-center justify-between border-b bg-surface-background px-3;
  }

  .panel-header h2,
  .console-header h1,
  .results-header h2 {
    @apply text-xs font-semibold text-fg-primary;
  }

  .panel-header p {
    @apply truncate font-mono text-[10px] text-fg-muted;
  }

  .table-list {
    @apply min-h-0 flex-1 overflow-auto py-1;
  }

  .console-main {
    @apply flex min-w-0 flex-1 flex-col overflow-hidden;
  }

  .read-only-badge {
    @apply rounded-full bg-surface-muted px-2 py-0.5 text-[10px] font-medium text-fg-secondary;
  }

  .execution-status {
    @apply flex items-center gap-x-1 text-[10px] text-fg-secondary;
  }

  .shortcut-hint {
    @apply text-[10px] text-fg-muted;
  }

  .worksheet-pane {
    @apply flex min-h-0 flex-1 flex-col overflow-hidden bg-surface-background;
  }

  .editor-pane {
    @apply min-h-0 flex-1 overflow-hidden;
  }

  .worksheet-error {
    @apply flex max-h-32 flex-none items-start gap-x-2 overflow-auto border-l-4 border-destructive bg-destructive/15 px-3 py-2 text-xs text-fg-primary;
  }

  .results-resizer {
    @apply relative h-1 flex-none border-y bg-surface-muted;
  }

  .results-pane {
    @apply flex min-h-40 flex-none flex-col overflow-hidden bg-surface-background;
  }

  .results-title,
  .results-meta {
    @apply flex items-center gap-x-2;
  }

  .results-title span,
  .results-meta {
    @apply text-[10px] text-fg-secondary;
  }

  .results-table {
    @apply min-h-0 flex-1 overflow-hidden;
  }

  @media (max-width: 720px) {
    .table-browser {
      @apply absolute inset-y-0 left-0 z-30 shadow-lg;
    }
  }
</style>
