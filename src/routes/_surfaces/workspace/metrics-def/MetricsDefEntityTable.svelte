<script lang="ts">
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
    "rounded border border-gray-200 overflow-auto flex-1 w-fit";
  const h4Class =
    "text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 font-bold uppercase align-middle flex-none";
</script>

<div class="metrics-def-section w-fit">
  <div class="flex flex-row pt-5 pb-3">
    <h4 class={h4Class}>{label}</h4>
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
    max-height: 50%;
    overflow-y: hidden;
    display: flex;
    flex-direction: column;
  }
</style>
