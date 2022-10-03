<script lang="ts">
  import type { ColumnConfig, CellConfig } from "./ColumnConfig";
  export let columnConfig: ColumnConfig<CellConfig>;
  export let index: number;
  export let row: any;
</script>

<svelte:component
  this={columnConfig.cellRenderer.component}
  {index}
  {row}
  {columnConfig}
/>
