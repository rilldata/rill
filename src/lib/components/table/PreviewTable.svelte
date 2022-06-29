<script lang="ts">
  /**
   * PreviewTable.svelte
   * Use this component to drop into the application.
   * Its goal it so utilize all of the other container components
   * and provide the interactions needed to do things with the table.
   */
  import { Table, TableRow, TableCell } from "$lib/components/table/";
  import PreviewTableHeader from "./PreviewTableHeader.svelte";
  import TableHeader from "./TableHeader.svelte";

  interface ColumnName {
    name: string;
    type: string;
  }

  export let columnNames: ColumnName[];
  export let rows: any[];

  let selectedColumns = [];

  let activeIndex;

  function columnIsPinned(name, selectedCols) {
    return selectedCols.map((column) => column.name).includes(name);
  }

  function togglePin(name, type, selectedCols) {
    // if column is already pinned, remove.
    if (columnIsPinned(name, selectedCols)) {
      selectedColumns = [
        ...selectedCols.filter((column) => column.name !== name),
      ];
    } else {
      selectedColumns = [...selectedCols, { name, type }];
    }
  }
</script>

<div class="flex relative">
  <Table>
    <!-- headers -->
    <TableRow>
      <TableHeader position="top-left">#</TableHeader>
      {#each columnNames as { name, type } (name)}
        {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
        <PreviewTableHeader
          {name}
          {type}
          pinned={thisColumnIsPinned}
          on:pin={() => {
            togglePin(name, type, selectedColumns);
          }}
        />
      {/each}
    </TableRow>
    <!-- values -->
    {#each rows as row, index}
      <TableRow hovered={activeIndex === index && activeIndex !== undefined}>
        <TableHeader position="left">{index + 1}</TableHeader>
        {#each columnNames as { name, type } (index + name)}
          <TableCell
            {name}
            {type}
            value={row[name]}
            isNull={row[name] === null}
          />
        {/each}
      </TableRow>
    {/each}
  </Table>

  {#if selectedColumns.length}
    <div
      class="sticky right-0 z-20 bg-white border border-l-4 border-y-0 border-r-0 border-gray-300"
    >
      <Table>
        <TableRow>
          {#each selectedColumns as { name, type } (name)}
            {@const thisColumnIsPinned = columnIsPinned(name, selectedColumns)}
            <PreviewTableHeader
              {name}
              {type}
              pinned={thisColumnIsPinned}
              on:pin={() => {
                togglePin(name, type, selectedColumns);
              }}
            />
          {/each}
        </TableRow>
        {#each rows as row, index}
          <TableRow
            hovered={activeIndex === index && activeIndex !== undefined}
          >
            {#each selectedColumns as { name, type }}
              <TableCell
                {name}
                {type}
                {index}
                isNull={row[name] === null}
                value={row[name]}
              />
            {/each}
          </TableRow>
        {/each}
      </Table>
    </div>
  {/if}
</div>
