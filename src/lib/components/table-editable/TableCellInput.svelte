<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import ErrorIcon from "$lib/components/icons/CrossIcon.svelte";
  import WarningIcon from "$lib/components/icons/WarningIcon.svelte";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  export let column: ColumnConfig;
  export let index = undefined;
  export let row: EntityRecord;

  let value = row[column.name];

  const dispatch = createEventDispatcher();

  let inputElt: HTMLInputElement;

  let editing = false;
  const onchangeHandler = (evt) => {
    stopEditing();
    dispatch("change", {
      value: evt.target.value,
      name: column.name,
      index,
    });
  };
  const startEditing = () => {
    editing = true;
    inputElt.focus();
  };
  const stopEditing = () => {
    editing = false;
    inputElt.blur();
  };
  // FIXME: validation is business logic that should be handled in
  // application state management, NOT in the component.
  $: validation = column.validation
    ? column.validation(row, row[column.name])
    : ValidationState.OK;
</script>

<td
  class="py-2 px-4 border border-gray-200 hover:bg-gray-200"
  style={editing
    ? "cursor:text; background:white; outline:1px solid #aaa;"
    : ""}
  on:click={startEditing}
>
  <div class="flex flex-row">
    <input
      bind:this={inputElt}
      id="model-title-input"
      class="table-input w-full text-ellipsis bg-inherit font-normal"
      on:input={startEditing}
      on:focus={startEditing}
      on:blur={stopEditing}
      on:change={onchangeHandler}
      value={value ?? ""}
    />

    {#if validation === ValidationState.ERROR}
      <ErrorIcon />
    {:else if validation === ValidationState.WARNING}
      <WarningIcon />
    {/if}
  </div>
</td>

<style>
  .table-input {
    cursor: default;
  }
  .table-input:focus-visible {
    border: none;
    outline: none;
    cursor: text;
  }
</style>
