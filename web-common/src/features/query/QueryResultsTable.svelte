<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import type {
    V1StructType,
    V1QueryResolverResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { prettyPrintType } from "./query-utils";

  export let schema: V1StructType | null;
  export let data: V1QueryResolverResponseDataItem[] | null;
  export let hasExecuted = false;

  $: columns = schemaToColumns(schema);

  function schemaToColumns(
    schema: V1StructType | null,
  ): VirtualizedTableColumns[] {
    if (!schema?.fields) return [];
    return schema.fields.map((field) => ({
      name: field.name ?? "",
      type: prettyPrintType(field.type?.code),
    }));
  }
</script>

{#if data && data.length > 0 && columns.length > 0}
  <PreviewTable rows={data} columnNames={columns} name="query-results" />
{:else if hasExecuted}
  <div class="empty-state">
    <p class="text-fg-secondary text-sm">No rows returned</p>
  </div>
{:else}
  <div class="empty-state">
    <p class="text-fg-secondary text-sm">Run query to view results</p>
  </div>
{/if}

<style lang="postcss">
  .empty-state {
    @apply flex items-center justify-center size-full;
  }
</style>
