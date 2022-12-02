<script lang="ts">
  import type { EntityRecord } from "@rilldata/web-local/lib/temp/entity";
  import { onMount } from "svelte";

  import type { ColumnConfig, CellConfigSelector } from "./ColumnConfig";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import AlertCircle from "../icons/AlertCircle.svelte";

  export let columnConfig: ColumnConfig<CellConfigSelector>;
  export let index: number;
  export let row: EntityRecord;

  const placeholderLabel = columnConfig.cellRenderer.placeholderLabel;
  $: options = columnConfig.cellRenderer.options;

  let value = undefined;
  let valueIsInvalid = false;
  let selectStyle = "";
  const greyedOut = " color: rgb(107, 114, 128);";

  $: if (row[columnConfig.name]) {
    // if there is an actual value, use it
    value = row[columnConfig.name];
    valueIsInvalid =
      options.find((selectorOption) => selectorOption.value == value) ===
      undefined;
    selectStyle = valueIsInvalid ? greyedOut : "";
  } else if (placeholderLabel) {
    // if there no actual value, use the placeholder label
    value = "__PLACEHOLDER_VALUE__";
    selectStyle = greyedOut + " font-style: italic;";
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

  let tdElt: HTMLTableCellElement;
  let tdHeight: number = undefined;
  onMount(() => {
    tdHeight = tdElt.getBoundingClientRect().height;
  });
  const ALERT_CIRCLE_SIZE_PX = 20;
  $: alertCircleTop = tdHeight / 2 - ALERT_CIRCLE_SIZE_PX / 2;
</script>

<td
  bind:this={tdElt}
  class="py-2 px-4 border border-gray-200 hover:bg-gray-200"
  style="position:relative"
>
  <select
    class="table-select bg-transparent w-full"
    on:change={onchangeHandler}
    style={selectStyle}
    {value}
  >
    {#if placeholderLabel}
      <option value="__PLACEHOLDER_VALUE__" disabled selected hidden
        >{placeholderLabel}</option
      >
    {/if}
    {#if valueIsInvalid}
      <option {value} disabled selected hidden>{value}</option>
    {/if}

    {#each options as option}
      <option value={option.value}>{option.label}</option>
    {/each}
  </select>

  {#if valueIsInvalid && tdHeight}
    <Tooltip location="top" alignment="middle" distance={16}>
      <div style={`position:absolute; top:${alertCircleTop}px; right:40px`}>
        <AlertCircle size={ALERT_CIRCLE_SIZE_PX + "px"} color={"#ef4444"} />
      </div>

      <TooltipContent slot="tooltip-content">
        {columnConfig.cellRenderer.invalidOptionMessage}
      </TooltipContent>
    </Tooltip>
  {/if}
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
