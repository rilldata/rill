<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import type {
    V1StructType,
    V1QueryResolverResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";

  export let schema: V1StructType | null;
  export let data: V1QueryResolverResponseDataItem[] | null;

  $: columns = schemaToColumns(schema);

  function schemaToColumns(
    schema: V1StructType | null,
  ): VirtualizedTableColumns[] {
    if (!schema?.fields) return [];
    return schema.fields.map((field) => ({
      name: field.name ?? "",
      type: typeCodeToString(field.type?.code),
    }));
  }

  function typeCodeToString(code: string | undefined): string {
    if (!code) return "UNKNOWN";
    // Strip the "CODE_" prefix for display
    return code.replace(/^CODE_/, "");
  }
</script>

{#if data && columns.length > 0}
  <PreviewTable rows={data} columnNames={columns} name="query-results" />
{:else}
  <div class="empty-state">
    <p class="text-fg-secondary text-sm">
      Run a query to see results
    </p>
  </div>
{/if}

<style lang="postcss">
  .empty-state {
    @apply flex items-center justify-center size-full;
  }
</style>
