<script lang="ts">
  import { onMount } from "svelte";

  import EditableTable from "../../table-editable/EditableTable.svelte";
  import AddIcon from "../../icons/Add.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";

  export let addEntityHandler;
  export let updateEntityHandler;
  export let deleteEntityHandler;
  export let rows;
  export let columnNames;
  export let tooltipText;
  export let addButtonId;
  export let label;
  export let resizeCallback;

  let sectionHeaderContainer;
  let sectionHeaderContainerHeight;
  onMount(() => {
    sectionHeaderContainerHeight = sectionHeaderContainer.clientHeight;
  });
</script>

<div class="metrics-def-section w-fit">
  <div class="flex flex-row pt-6 pb-3" bind:this={sectionHeaderContainer}>
    <h4
      class="text-ellipsis overflow-hidden whitespace-nowrap text-gray-400 ui-copy-strong uppercase align-middle flex-none"
    >
      {label}
    </h4>
    <div class="align-middle pl-5">
      <ContextButton id={addButtonId} {tooltipText} on:click={addEntityHandler}>
        <AddIcon />
      </ContextButton>
    </div>
  </div>
  <div class="rounded border border-gray-200 overflow-auto flex-1 max-w-[100%]">
    <EditableTable
      {rows}
      {columnNames}
      on:change={updateEntityHandler}
      on:delete={deleteEntityHandler}
      on:tableResize={(evt) =>
        resizeCallback(evt.detail + sectionHeaderContainerHeight)}
    />
  </div>
</div>

<style>
  .metrics-def-section {
    overflow: hidden;
    display: flex;
    flex-direction: column;
    max-width: 100%;
    height: 100%;
  }
</style>
