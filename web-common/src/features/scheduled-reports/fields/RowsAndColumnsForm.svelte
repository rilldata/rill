<script lang="ts">
  import FieldList from "@rilldata/web-common/features/scheduled-reports/fields/FieldList.svelte";
  import { getFieldsForExplore } from "@rilldata/web-common/features/scheduled-reports/fields/selectors.ts";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import type { ValidationErrors } from "sveltekit-superforms";

  export let rows: string[];
  export let columns: string[];
  export let columnErrors: ValidationErrors<ReportValues>["columns"];
  export let instanceId: string;
  export let exploreName: string;

  $: selectedFields = new Set([...rows, ...columns]);

  $: fieldsForExplore = getFieldsForExplore(instanceId, exploreName);
  $: ({ displayMap, allowedRows, allowedColumns } = $fieldsForExplore ?? {});

  $: hasSomeRow = rows.length > 0;
  // If there are some rows then we need force measures last.
  // So we need to disable drag drop and sort existing columns.
  $: disableColumnDragDrop = hasSomeRow;
  $: if (hasSomeRow) sortColumnsMeasuresLast();

  $: hasColumnErrors = Boolean(columnErrors?._errors?.length);

  function sortColumnsMeasuresLast() {
    const newColumns = [...columns];
    newColumns.sort((a, b) => {
      const aIsMeasure = displayMap[a]?.type === "measure";
      const bIsMeasure = displayMap[b]?.type === "measure";
      return !aIsMeasure && bIsMeasure ? -1 : 1;
    });
    columns = newColumns;
  }

  function handleColumnUpdate(newColumns: string[]) {
    columns = newColumns;
    if (hasSomeRow) sortColumnsMeasuresLast();
  }
</script>

{#if displayMap}
  <FieldList
    bind:fields={rows}
    allowedFields={allowedRows.filter((r) => !selectedFields.has(r))}
    {displayMap}
    label="Rows"
    onUpdate={(newRows) => (rows = newRows)}
  >
    <div slot="empty-fields" class="text-fg-secondary">No rows selected</div>
  </FieldList>

  <FieldList
    bind:fields={columns}
    allowedFields={allowedColumns.filter((r) => !selectedFields.has(r))}
    {displayMap}
    label="Columns"
    disableDragDrop={disableColumnDragDrop}
    onUpdate={handleColumnUpdate}
  >
    <div
      slot="empty-fields"
      class={hasColumnErrors ? "text-red-600" : "text-fg-secondary"}
    >
      Must select one column
    </div>
  </FieldList>
{/if}
