<script lang="ts">
  import { onMount } from "svelte";

  import PreviewTable from "$lib/components/table-editable/PreviewTable.svelte";
  import AddIcon from "$lib/components/icons/Add.svelte";
  import ContextButton from "$lib/components/column-profile/ContextButton.svelte";

  export let addEntityHandler;
  export let updateEntityHandler;
  export let deleteEntityHandler;
  export let rows;
  export let columnNames;
  export let tooltipText;
  export let addButtonId;
  export let label;

  const tableContainerDivClass =
    "rounded border border-gray-200 overflow-auto flex-1 max-w-[100%]";

  let sectionHeaderContainer;
  let sectionHeaderContainerHeight;

  // FIXME: table rows currently 37.2px tall, table header is 43,
  // but need a better way to calculate this from actual elements
  const TABLE_HEADER_HEIGHT = 43;
  const TABLE_ROW_HEIGHT = 37.2;
  const MIN_ROWS_SHOWN = 2;
  $: tableHeightPx = Math.round(
    Math.min(MIN_ROWS_SHOWN, rows.length) * TABLE_ROW_HEIGHT +
      TABLE_HEADER_HEIGHT +
      sectionHeaderContainerHeight
  );
  let sectionContainerStyles = "";
  $: sectionContainerStyles = `min-height: ${tableHeightPx}px;`;

  const entityTableHeaderClass =
    "text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 font-bold uppercase align-middle flex-none";

  onMount(() => {
    const observer = new ResizeObserver(() => {
      sectionHeaderContainerHeight = sectionHeaderContainer.clientHeight;
    });
    observer.observe(sectionHeaderContainer);
    return () => observer.unobserve(sectionHeaderContainer);
  });
</script>

<div class="metrics-def-section w-fit" style={sectionContainerStyles}>
  <div class="flex flex-row pt-5 pb-3" bind:this={sectionHeaderContainer}>
    <h4 class={entityTableHeaderClass}>
      {label}
    </h4>
    <div class="align-middle pl-5">
      <ContextButton id={addButtonId} {tooltipText} on:click={addEntityHandler}>
        <AddIcon />
      </ContextButton>
    </div>
  </div>
  <div class={tableContainerDivClass}>
    <PreviewTable
      tableConfig={{ enableAdd: false }}
      {rows}
      {columnNames}
      on:change={updateEntityHandler}
      on:delete={deleteEntityHandler}
    />
  </div>
</div>

<style>
  .metrics-def-section {
    overflow: hidden;
    display: flex;
    flex-direction: column;
    max-width: 100%;
  }
</style>
