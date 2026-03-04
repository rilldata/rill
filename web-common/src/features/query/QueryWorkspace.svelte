<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import { runtime } from "../../runtime-client/runtime-store";
  import ConnectorSelector from "./ConnectorSelector.svelte";
  import QueryEditor from "./QueryEditor.svelte";
  import QueryResultsInspector from "./QueryResultsInspector.svelte";
  import QueryResultsTable from "./QueryResultsTable.svelte";
  import { createQueryConsole } from "./query-store";

  const WORKSPACE_KEY = "__query_console__";

  const queryConsole = createQueryConsole();

  $: ({ instanceId } = $runtime);

  $: workspace = workspaces.get(WORKSPACE_KEY);
  $: tableVisible = workspace.table.visible;

  // Derived state from query console
  let sql = "";
  let connector = "";
  let limit = 100;

  $: schema = queryConsole.schema;
  $: data = queryConsole.data;
  $: rowCount = queryConsole.rowCount;

  function handleRun() {
    queryConsole.setSql(sql);
    queryConsole.setConnector(connector);
    queryConsole.setLimit(limit);
    queryConsole.executeQuery(instanceId);
  }

  function handleEditorChange(e: CustomEvent<string>) {
    sql = e.detail;
  }

  function handleConnectorChange(newConnector: string) {
    connector = newConnector;
  }

  function handleLimitChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const parsed = parseInt(target.value, 10);
    if (!isNaN(parsed) && parsed > 0) {
      limit = parsed;
    }
  }
</script>

<WorkspaceContainer>
  <header slot="header" class="query-header">
    <div class="flex items-center gap-x-3 px-4 py-2">
      <h2 class="text-lg font-semibold flex-none">Query</h2>

      <div class="flex items-center gap-x-2 flex-none">
        <span class="text-xs text-fg-secondary">Connector</span>
        <ConnectorSelector
          value={connector}
          onChange={handleConnectorChange}
        />
      </div>

      <div class="flex items-center gap-x-2 flex-none">
        <span class="text-xs text-fg-secondary">Limit</span>
        <input
          type="number"
          class="limit-input"
          value={limit}
          min="1"
          max="10000"
          on:change={handleLimitChange}
        />
      </div>

      <div class="flex items-center gap-x-2 ml-auto flex-none">
        {#if $queryConsole.isExecuting}
          <div class="flex items-center gap-x-2 text-xs text-fg-secondary">
            <Spinner size="14px" status={EntityStatus.Running} />
            Running...
          </div>
        {:else if $queryConsole.result}
          <span class="text-xs text-fg-secondary">
            {formatInteger($rowCount)} {$rowCount === 1 ? "row" : "rows"}
            {#if $queryConsole.executionTimeMs !== null}
              in {$queryConsole.executionTimeMs < 1000
                ? `${$queryConsole.executionTimeMs}ms`
                : `${($queryConsole.executionTimeMs / 1000).toFixed(1)}s`}
            {/if}
          </span>
        {/if}

        <Button
          type="primary"
          onClick={handleRun}
          disabled={$queryConsole.isExecuting || !sql.trim()}
        >
          Run
        </Button>
      </div>
    </div>
  </header>

  <div
    slot="body"
    class="editor-pane size-full overflow-hidden flex flex-col"
  >
    <WorkspaceEditorContainer error={$queryConsole.error ?? undefined}>
      <QueryEditor
        on:run={handleRun}
        on:change={handleEditorChange}
      />
    </WorkspaceEditorContainer>

    {#if $tableVisible}
      <WorkspaceTableContainer filePath={WORKSPACE_KEY}>
        {#if $queryConsole.isExecuting}
          <div class="size-full flex items-center justify-center">
            <Spinner size="1.5em" status={EntityStatus.Running} />
          </div>
        {:else}
          <QueryResultsTable schema={$schema} data={$data} />
        {/if}
      </WorkspaceTableContainer>
    {/if}
  </div>

  <QueryResultsInspector
    slot="inspector"
    filePath={WORKSPACE_KEY}
    schema={$schema}
    rowCount={$rowCount}
    executionTimeMs={$queryConsole.executionTimeMs}
  />
</WorkspaceContainer>

<style lang="postcss">
  .query-header {
    @apply border-b bg-surface-base;
  }

  .limit-input {
    @apply w-20 h-6 px-2 text-xs rounded border bg-input text-fg-primary;
  }

  .limit-input:focus {
    @apply outline-none ring-2 ring-primary-100;
  }

  /* Hide number input spinners */
  .limit-input::-webkit-outer-spin-button,
  .limit-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  .limit-input[type="number"] {
    -moz-appearance: textfield;
  }
</style>
