<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import ErrorIcon from "$lib/components/icons/CrossIcon.svelte";
  import WarningIcon from "$lib/components/icons/WarningIcon.svelte";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  export let columnConfig: ColumnConfig;
  export let index: number;
  export let row: EntityRecord;
  export let value: string;
  // export let selectionHandler: (evt: Event) => void;

  // let value = row[column.name];

  const onchangeHandler = (evt) => {
    columnConfig.onchange(index, columnConfig.name, evt.target.value);
  };

  $: options = columnConfig.options;
  $: initialValue = options.length > 0 ? options[0] : undefined;

  value = value ?? "__DEFAULT_VALUE__";
</script>

<td class="py-2 px-4 border border-gray-200 hover:bg-gray-200">
  <select
    class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
    style="background-color: #FFF; width:18em;"
    {value}
    on:change={onchangeHandler}
  >
    <option value="__DEFAULT_VALUE__" disabled selected hidden
      >select a timestamp...</option
    >
    {#each options as option}
      <option value={option}>{option}</option>
    {/each}
  </select>
</td>

<!-- <style>
  .table-input {
    cursor: default;
  }
  .table-input:focus-visible {
    border: none;
    outline: none;
    cursor: text;
  }
</style> -->
