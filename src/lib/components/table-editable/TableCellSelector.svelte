<script lang="ts">
  import type {
    ColumnConfig,
    CellConfigSelector,
  } from "$lib/components/table-editable/ColumnConfig";
  import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  export let columnConfig: ColumnConfig<CellConfigSelector>;
  export let index: number;
  export let row: EntityRecord;

  const placeholderLabel = columnConfig.cellRenderer.placeholderLabel;

  let value = undefined;
  let placeholderStyles = "";
  $: if (row[columnConfig.name]) {
    // if there is an actual value, use it
    value = row[columnConfig.name];
  } else if (placeholderLabel) {
    // if there no actual value, use the placeholder label
    value = "__PLACEHOLDER_VALUE__";
    placeholderStyles = "font-style: italic; color: rgb(107, 114, 128);";
  } else {
    // if there is no placeholder label and no actual value, use the first option in the options list
    value = columnConfig.cellRenderer.options[0].value;
  }

  const onchangeHandler = (evt) => {
    columnConfig.cellRenderer.onchange(
      index,
      columnConfig.name,
      evt.target.value
    );
  };

  $: options = columnConfig.cellRenderer.options;
</script>

<td class="py-2 px-4 border border-gray-200 hover:bg-gray-200">
  <select
    class="table-select bg-transparent"
    style={placeholderStyles}
    {value}
    on:change={onchangeHandler}
  >
    {#if placeholderLabel}
      <option value="__PLACEHOLDER_VALUE__" disabled selected hidden
        >{placeholderLabel}</option
      >
    {/if}

    {#each options as option}
      <option value={option.value}>{option.label}</option>
    {/each}
  </select>
</td>

<style>
  .table-select {
    cursor: default;
  }
  .table-select:focus-visible {
    border: none;
    outline: none;
  }
</style>
