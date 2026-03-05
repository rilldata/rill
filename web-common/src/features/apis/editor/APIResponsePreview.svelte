<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import WorkspaceError from "@rilldata/web-common/components/WorkspaceError.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import IconSwitcher from "@rilldata/web-common/components/forms/IconSwitcher.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { TableIcon, BracesIcon } from "lucide-svelte";

  export let response: unknown[] | null;
  export let error: string | null;
  export let isLoading: boolean;
  export let apiName: string;

  type ViewMode = "table" | "json";
  let viewMode: ViewMode = "table";

  const viewModeOptions = [
    { id: "table", Icon: TableIcon, tooltip: "Table view" },
    { id: "json", Icon: BracesIcon, tooltip: "JSON view" },
  ];

  function handleViewModeChange(id: string) {
    if (id === "table" || id === "json") {
      viewMode = id;
    }
  }

  $: columns = extractColumns(response);

  function extractColumns(data: unknown[] | null): VirtualizedTableColumns[] {
    if (!data || data.length === 0) return [];

    const firstRow = data[0];
    if (typeof firstRow !== "object" || firstRow === null) {
      return [{ name: "value", type: "VARCHAR" }];
    }

    const keys = Object.keys(firstRow);
    const sampleSize = Math.min(data.length, 10);

    return keys.map((key) => ({
      name: key,
      type: inferTypeFromSample(data, key, sampleSize),
    }));
  }

  function inferTypeFromSample(
    rows: unknown[],
    key: string,
    sampleSize: number,
  ): string {
    for (let i = 0; i < sampleSize; i++) {
      const row = rows[i];
      if (typeof row !== "object" || row === null) continue;
      const value = (row as Record<string, unknown>)[key];
      if (value !== null && value !== undefined) {
        return inferType(value);
      }
    }
    return "VARCHAR";
  }

  function inferType(value: unknown): string {
    if (value === null || value === undefined) return "VARCHAR";
    if (typeof value === "number") {
      return Number.isInteger(value) ? "INTEGER" : "DOUBLE";
    }
    if (typeof value === "boolean") return "BOOLEAN";
    if (typeof value === "object") return "JSON";
    return "VARCHAR";
  }

  function normalizeRows(data: unknown[] | null): Record<string, unknown>[] {
    if (!data) return [];

    return data.map((item) => {
      if (typeof item !== "object" || item === null) {
        return { value: item };
      }
      return item as Record<string, unknown>;
    });
  }

  $: normalizedRows = normalizeRows(response);
</script>

<div class="response-preview-wrapper">
  <div class="response-header">
    <span class="header-label">Response Preview</span>
    {#if response && response.length > 0}
      <IconSwitcher
        fields={viewModeOptions}
        selected={viewMode}
        onClick={handleViewModeChange}
        small
      />
    {/if}
  </div>
  <div class="response-content">
    {#if isLoading}
      <ReconcilingSpinner />
    {:else if error}
      <WorkspaceError message={error} />
    {:else if !response}
      <div class="empty-state">Click "Test API" to see the response</div>
    {:else if response.length === 0}
      <div class="empty-state">API returned an empty response</div>
    {:else if viewMode === "json"}
      <pre class="json-view">{JSON.stringify(response, null, 2)}</pre>
    {:else}
      <PreviewTable
        rows={normalizedRows}
        columnNames={columns}
        name={apiName}
      />
    {/if}
  </div>
</div>

<style lang="postcss">
  .response-preview-wrapper {
    @apply flex flex-col h-full;
  }

  .response-header {
    @apply flex items-center justify-between px-3 py-2 border-b bg-surface-muted;
  }

  .header-label {
    @apply text-xs font-medium text-fg-secondary uppercase;
  }

  .response-content {
    @apply flex-1 overflow-auto;
  }

  .empty-state {
    @apply flex items-center justify-center h-full text-fg-muted text-sm;
  }

  .json-view {
    @apply p-4 text-sm font-mono whitespace-pre-wrap break-words;
    @apply bg-surface-background text-fg-primary;
  }
</style>
