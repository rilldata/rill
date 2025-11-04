<script lang="ts">
  import { getContext } from "svelte";
  import ColumnHeader from "../core/ColumnHeader.svelte";
  import type {
    VirtualizedTableColumns,
    VirtualizedTableConfig,
  } from "../types";

  export let columns: VirtualizedTableColumns[];
  export let pinnedColumns: VirtualizedTableColumns[] = [];
  export let virtualColumnItems;
  export let noPin = false;
  export let showDataIcon = false;
  export let sortByMeasure: string | null = null;
  export let onClickColumn: (columnName: string) => void = () => {};
  export let onPin: (column: VirtualizedTableColumns) => void = () => {};

  const config: VirtualizedTableConfig = getContext("config");

  const getColumnHeaderProps = (header) => {
    const column = columns[header.index];
    const name = column.label || column.name;
    const isEnableResizeDefined = "enableResize" in column;
    const enableResize = isEnableResizeDefined ? column.enableResize : true;
    const enableSorting =
      "enableSorting" in column
        ? column.enableResize
        : config.table === "DimensionTable";
    return {
      name,
      enableResize,
      enableSorting,
      type: column.type,
      description: column.description || "",
      pinned: pinnedColumns.some((pinCol) => pinCol.name === column.name),
      isSelected: sortByMeasure === column.name,
      sorted: column.sorted,
    };
  };
</script>

<div class="w-full sticky top-0 z-10">
  {#each virtualColumnItems as header (header.key)}
    {@const props = getColumnHeaderProps(header)}
    <ColumnHeader
      {...props}
      {header}
      {noPin}
      {showDataIcon}
      onPin={() => {
        onPin(columns[header.index]);
      }}
      onClickColumn={() => {
        onClickColumn(columns[header.index].name);
      }}
    />
  {/each}
</div>
