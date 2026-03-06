<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import { createEventDispatcher } from "svelte";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import ConnectorSelector from "./ConnectorSelector.svelte";
  import QueryEditor from "./QueryEditor.svelte";
  import QueryResultsTable from "./QueryResultsTable.svelte";
  import type { NotebookStore } from "./query-store";

  const dispatch = createEventDispatcher<{ focus: void; run: void }>();

  export let cellId: string;
  export let notebook: NotebookStore;
  export let cellCount: number;

  const runtimeClient = useRuntimeClient();

  let editorRef: QueryEditor;
  let cellHeight = 450;

  /** Called externally (e.g. from data explorer) to set the editor content */
  export function setEditorContent(text: string) {
    editorRef?.setContent(text);
  }

  /** Inserts text at the current cursor position */
  export function insertAtCursor(text: string) {
    editorRef?.insertAtCursor(text);
  }

  $: cell = $notebook.cells.find((c) => c.id === cellId);
  $: isFocused = $notebook.focusedCellId === cellId;
  $: canDelete = cellCount > 1;

  $: schema = cell?.result?.schema ?? null;
  $: data = cell?.result?.data ?? null;
  $: rowCount = (cell?.result?.data?.length || cell?.lastRowCount) ?? 0;
  $: hasExecuted = cell?.hasExecuted ?? false;
  $: hasSql = (cell?.sql ?? "").trim().length > 0;

  function runQuery(sqlOverride?: string) {
    if (!cell || cell.isExecuting) return;
    notebook.setFocusedCell(cellId);
    notebook.executeCellQuery(cellId, runtimeClient, sqlOverride);
    dispatch("run");
  }

  function handleRun(e: CustomEvent<{ selectedText?: string }>) {
    runQuery(e.detail?.selectedText);
  }

  function handleRunButton() {
    runQuery(editorRef?.getSelectedText());
  }

  function handleChange(e: CustomEvent<string>) {
    notebook.setCellSql(cellId, e.detail);
  }

  function handleConnectorChange(newConnector: string) {
    notebook.setCellConnector(cellId, newConnector);
  }

  function handleLimitChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const val = target.value.trim();
    if (val === "") {
      notebook.setCellLimit(cellId, undefined);
    } else {
      const parsed = Number(val);
      if (Number.isFinite(parsed) && parsed > 0) {
        notebook.setCellLimit(cellId, parsed);
      }
    }
  }

  function handleFocus() {
    notebook.setFocusedCell(cellId);
    dispatch("focus");
    if (!cell?.collapsed) {
      editorRef?.focus();
    }
  }
</script>

{#if cell}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="query-cell" class:focused={isFocused} on:click={handleFocus}>
    <!-- Cell Header -->
    <div
      class="cell-header"
      on:click|stopPropagation
      role="toolbar"
      tabindex="-1"
    >
      <button
        class="collapse-toggle"
        on:click={() => notebook.toggleCellCollapsed(cellId)}
        aria-label={cell.collapsed ? "Expand cell" : "Collapse cell"}
      >
        <CaretDownIcon
          className="transform transition-transform {cell.collapsed
            ? '-rotate-90'
            : 'rotate-0'}"
          size="14px"
        />
      </button>

      <ConnectorSelector
        id="connector-{cellId}"
        value={cell.connector}
        onChange={handleConnectorChange}
      />

      <div class="flex items-center gap-x-1.5 flex-none">
        <span class="text-[11px] text-fg-secondary">Limit</span>
        <input
          type="number"
          class="limit-input"
          value={cell.limit ?? ""}
          placeholder="None"
          min="1"
          on:change={handleLimitChange}
        />
      </div>

      {#if cell.limit === undefined}
        <span class="limit-warning">
          Server default (10,000 rows) applies. Adjustable via
          rill.interactive_sql_row_limit.
        </span>
      {/if}

      <div class="flex items-center gap-x-2 ml-auto flex-none">
        {#if cell.isExecuting}
          <div
            class="flex items-center gap-x-1.5 text-[11px] text-fg-secondary"
          >
            <Spinner size="12px" status={EntityStatus.Running} />
            Running...
          </div>
        {:else if cell.result}
          <span class="text-[11px] text-fg-secondary">
            {formatInteger(rowCount)}
            {rowCount === 1 ? "row" : "rows"}
            {#if cell.executionTimeMs !== null}
              in {cell.executionTimeMs < 1000
                ? `${cell.executionTimeMs}ms`
                : `${(cell.executionTimeMs / 1000).toFixed(1)}s`}
            {/if}
          </span>
        {/if}

        <Button
          type="primary"
          small
          onClick={handleRunButton}
          disabled={cell.isExecuting || !hasSql}
        >
          Run
        </Button>

        {#if canDelete}
          <button
            class="delete-button"
            on:click|stopPropagation={() => notebook.removeCell(cellId)}
            aria-label="Delete cell"
          >
            ×
          </button>
        {/if}
      </div>
    </div>

    <!-- Editor + Results -->
    {#if !cell.collapsed}
      <div class="cell-body" style:height="{cellHeight}px">
        <div class="editor-pane">
          <WorkspaceEditorContainer error={cell.error ?? undefined}>
            <QueryEditor
              bind:this={editorRef}
              initialValue={cell.sql}
              on:run={handleRun}
              on:change={handleChange}
            />
          </WorkspaceEditorContainer>
        </div>

        {#if cell.result || cell.isExecuting}
          <div class="cell-results">
            <div class="results-wrapper">
              {#if cell.isExecuting}
                <div class="size-full flex items-center justify-center">
                  <Spinner size="1.5em" status={EntityStatus.Running} />
                </div>
              {:else}
                <QueryResultsTable {schema} {data} {hasExecuted} />
              {/if}
            </div>
          </div>
        {/if}

        <div class="cell-resize-handle">
          <Resizer
            absolute={false}
            direction="NS"
            side="top"
            min={120}
            max={900}
            bind:dimension={cellHeight}
          />
        </div>
      </div>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .query-cell {
    @apply border rounded bg-surface-background;
  }

  .query-cell.focused {
    @apply ring-1 ring-primary-300;
  }

  .cell-header {
    @apply flex items-center gap-x-2 px-3 py-1.5;
    @apply border-b bg-surface-subtle;
  }

  .collapse-toggle {
    @apply flex-none p-0.5 rounded;
  }

  .collapse-toggle:hover {
    @apply bg-gray-200;
  }

  .limit-input {
    @apply w-16 h-6 px-2 text-[11px] rounded border bg-input text-fg-primary;
  }

  .limit-input:focus {
    @apply outline-none ring-1 ring-primary-100;
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

  .limit-warning {
    @apply text-[10px] text-yellow-600 flex-none;
  }

  .cell-body {
    @apply flex flex-col overflow-hidden;
  }

  .editor-pane {
    @apply flex-none overflow-hidden;
    height: 40%;
    min-height: 60px;
  }

  .cell-results {
    @apply flex-1 min-h-0 flex flex-col;
  }

  .results-wrapper {
    @apply relative w-full overflow-hidden border-t h-full;
  }

  .cell-resize-handle {
    @apply flex-none h-2 relative cursor-ns-resize;
    @apply bg-surface-subtle border-t hover:bg-primary-100;
    transition: background-color 0.15s;
  }

  .delete-button {
    @apply text-fg-secondary text-lg leading-none px-1 rounded;
  }

  .delete-button:hover {
    @apply text-fg-primary bg-gray-200;
  }
</style>
