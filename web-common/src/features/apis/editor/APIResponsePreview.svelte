<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import WorkspaceError from "@rilldata/web-common/components/WorkspaceError.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";

  export let response: unknown[] | null;
  export let error: string | null;
  export let isLoading: boolean;
  export let apiName: string;

  $: columns = extractColumns(response);
  $: rows = response ?? [];

  function extractColumns(data: unknown[] | null): VirtualizedTableColumns[] {
    if (!data || data.length === 0) return [];

    const firstRow = data[0];
    if (typeof firstRow !== "object" || firstRow === null) {
      return [{ name: "value", type: "VARCHAR" }];
    }

    return Object.keys(firstRow).map((key) => ({
      name: key,
      type: inferType(firstRow[key as keyof typeof firstRow]),
    }));
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

{#if isLoading}
  <ReconcilingSpinner />
{:else if error}
  <WorkspaceError message={error} />
{:else if !response}
  <div class="flex items-center justify-center h-full text-fg-muted text-sm">
    Click "Test API" to see the response
  </div>
{:else if response.length === 0}
  <div class="flex items-center justify-center h-full text-fg-muted text-sm">
    API returned an empty response
  </div>
{:else}
  <PreviewTable rows={normalizedRows} columnNames={columns} name={apiName} />
{/if}
