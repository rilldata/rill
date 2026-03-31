<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
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
  import { downloadResultsAsCSV, downloadResultsAsJSON } from "./query-export";
  import { formatExecutionTime } from "./query-utils";

  const dispatch = createEventDispatcher<{ focus: void; run: void }>();

  export let cellId: string;
  export let notebook: NotebookStore;
  export let cellCount: number;

  const runtimeClient = useRuntimeClient();

  let editorRef: QueryEditor;
  let editorHeight = 180;
  let resultsHeight = 250;

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
  $: hasExecuted = cell?.hasExecuted ?? false;
  $: rowCount = hasExecuted
    ? (cell?.result?.data?.length ?? 0)
    : (cell?.lastRowCount ?? 0);
  $: hasSql = (cell?.sql ?? "").trim().length > 0;
  $: hasResults = (data?.length ?? 0) > 0 && (schema?.fields?.length ?? 0) > 0;

  function handleRunButton() {
    if (!cell || cell.isExecuting) return;
    notebook.setFocusedCell(cellId);
    const selected = editorRef?.getSelectedText();
    notebook.executeCellQuery(cellId, runtimeClient, selected);
    dispatch("run");
  }

  function handleStopButton() {
    notebook.cancelCellQuery(cellId);
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
      const parsed = parseInt(val, 10);
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
        {:else if cell.result && (hasExecuted || cell.lastRowCount)}
          <span class="text-[11px] text-fg-secondary">
            {formatInteger(rowCount)}
            {rowCount === 1 ? "row" : "rows"}
            {#if cell.executionTimeMs !== null}
              in {formatExecutionTime(cell.executionTimeMs)}
            {/if}
          </span>
        {/if}

        {#if cell.isExecuting}
          <Button type="destructive" small onClick={handleStopButton}>
            Stop
          </Button>
        {:else}
          <Button
            type="primary"
            small
            onClick={handleRunButton}
            disabled={!hasSql}
          >
            Run
          </Button>
        {/if}

        {#if hasResults}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <Button
                  {...props}
                  label="Export results"
                  type="secondary"
                  small
                >
                  <Export size="13px" />
                </Button>
              {/snippet}
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              <DropdownMenu.Item
                onclick={() => downloadResultsAsCSV(schema, data)}
              >
                Download as CSV
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onclick={() => downloadResultsAsJSON(schema, data)}
              >
                Download as JSON
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

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
      <div class="cell-body">
        <div class="editor-pane" style:height="{editorHeight}px">
          <WorkspaceEditorContainer>
            <QueryEditor
              bind:this={editorRef}
              initialValue={cell.sql}
              on:change={handleChange}
            />
          </WorkspaceEditorContainer>
        </div>

        {#if cell.error}
          <!-- svelte-ignore a11y-click-events-have-key-events -->
          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <div class="cell-error" on:click|stopPropagation>
            <CancelCircle className="text-destructive flex-none" />
            <span>{cell.error}</span>
          </div>
        {/if}

        <div class="resize-handle">
          <Resizer
            absolute={false}
            direction="NS"
            side="bottom"
            min={60}
            max={600}
            hang={false}
            bind:dimension={editorHeight}
          />
        </div>

        {#if cell.result || cell.isExecuting}
          <div class="cell-results" style:height="{resultsHeight}px">
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

          <div class="resize-handle">
            <Resizer
              absolute={false}
              direction="NS"
              side="bottom"
              min={80}
              max={800}
              hang={false}
              bind:dimension={resultsHeight}
            />
          </div>
        {/if}
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
    min-height: 60px;
  }

  .cell-error {
    @apply flex items-center gap-x-2 px-3 py-2 text-sm text-fg-primary;
    @apply border-l-4 border-destructive bg-destructive/15;
    @apply max-h-40 overflow-auto select-text;
  }

  .cell-results {
    @apply flex-none min-h-0 flex flex-col overflow-hidden;
  }

  .results-wrapper {
    @apply relative w-full overflow-hidden border-t h-full;
  }

  .resize-handle {
    @apply flex-none relative cursor-ns-resize border-t border-b;
    height: 5px;
  }

  .resize-handle:hover {
    @apply bg-primary-100;
  }

  .delete-button {
    @apply flex items-center justify-center text-fg-secondary text-lg px-1 rounded;
  }

  .delete-button:hover {
    @apply text-fg-primary bg-gray-200;
  }
</style>
