<script lang="ts">
  import FieldList from "@rilldata/web-common/features/scheduled-reports/fields/FieldList.svelte";
  import { getFieldsForExplore } from "@rilldata/web-common/features/scheduled-reports/fields/selectors.ts";

  export let rows: string[];
  export let columns: string[];
  export let instanceId: string;
  export let exploreName: string;

  $: selectedFields = new Set([...rows, ...columns]);

  $: fieldsForExplore = getFieldsForExplore(instanceId, exploreName);
  $: ({ displayMap, allowedRows, allowedColumns } = $fieldsForExplore);
</script>

<FieldList
  bind:fields={rows}
  allowedFields={allowedRows.filter((r) => !selectedFields.has(r))}
  {displayMap}
  label="Rows"
/>

<FieldList
  bind:fields={columns}
  allowedFields={allowedColumns.filter((r) => !selectedFields.has(r))}
  {displayMap}
  label="Columns"
/>
