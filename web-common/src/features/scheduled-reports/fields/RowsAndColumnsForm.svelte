<script lang="ts">
  import FieldList from "@rilldata/web-common/features/scheduled-reports/fields/FieldList.svelte";
  import { getFieldsForExplore } from "@rilldata/web-common/features/scheduled-reports/fields/selectors.ts";

  export let rows: string[];
  export let columns: string[];
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
  />

  <FieldList
    bind:fields={columns}
    allowedFields={allowedColumns.filter((r) => !selectedFields.has(r))}
    {displayMap}
    label="Columns"
    disableDragDrop={disableColumnDragDrop}
    onUpdate={handleColumnUpdate}
  />
{/if}
