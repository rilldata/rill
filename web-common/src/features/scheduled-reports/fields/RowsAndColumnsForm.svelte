<script lang="ts">
  import FieldList from "@rilldata/web-common/features/scheduled-reports/fields/FieldList.svelte";
  import { getFieldsForExplore } from "@rilldata/web-common/features/scheduled-reports/fields/selectors.ts";

  export let rows: string[];
  export let columns: string[];
  export let instanceId: string;
  export let exploreName: string;
  // Used to highlight missing columns in red. We only want to highlight when the user submits explicitly.
  export let didSubmit: boolean;

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
  >
    <div slot="empty-fields" class="text-gray-500">No selected rows</div>
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
      class={didSubmit ? "text-red-600" : "text-gray-500"}
    >
      Must select one column
    </div>
  </FieldList>
{/if}
