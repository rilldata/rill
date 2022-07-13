<script lang="ts">
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  export let columnConfig: ColumnConfig;
  export let index: number;
  export let row: any;
</script>

<svelte:component
  this={columnConfig.renderer}
  value={row[columnConfig.name]}
  {index}
  {row}
  {columnConfig}
/>
