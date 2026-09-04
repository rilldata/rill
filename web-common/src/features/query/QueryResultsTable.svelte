<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import type { V1QueryResolverResponse } from "@rilldata/web-common/runtime-client";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { formatDataType } from "./query-utils";

  let {
    result,
    hasExecuted,
    error = null,
  }: {
    result: V1QueryResolverResponse | null;
    hasExecuted: boolean;
    error?: string | null;
  } = $props();

  let rows = $derived(result?.data ?? []);
  let columns = $derived(
    (result?.schema?.fields ?? []).map(
      (field): VirtualizedTableColumns => ({
        name: field.name ?? "",
        type: formatDataType(field.type?.code),
      }),
    ),
  );
</script>

{#if rows.length > 0 && columns.length > 0}
  <PreviewTable {rows} columnNames={columns} name="sql-console-results" />
{:else if error}
  <div class="empty-state">
    <p>Query failed. Review the error above and try again.</p>
  </div>
{:else if hasExecuted}
  <div class="empty-state">
    <p>No rows returned</p>
  </div>
{:else}
  <div class="empty-state">
    <p>Run a query to view results</p>
  </div>
{/if}

<style lang="postcss">
  .empty-state {
    @apply flex size-full items-center justify-center text-sm text-fg-secondary;
  }
</style>
