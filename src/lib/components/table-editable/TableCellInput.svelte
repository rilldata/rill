<script lang="ts">
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import AlertCircle from "$lib/components/icons/AlertCircle.svelte";
  import AlertTriangle from "$lib/components/icons/AlertTriangle.svelte";

  import type {
    ColumnConfig,
    CellConfigInput,
  } from "$lib/components/table-editable/ColumnConfig";
  import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  // FIXME: this import below will be needed for typing
  // `(<MeasureDefinitionEntity>row)` once we have more detailed
  // validation messages
  // import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

  export let columnConfig: ColumnConfig<CellConfigInput>;
  export let index = undefined;
  export let row: EntityRecord;
  $: value = row[columnConfig.name];

  let inputElt: HTMLInputElement;

  let editing = false;
  const onchangeHandler = (evt) => {
    stopEditing();
    columnConfig.cellRenderer.onchange(
      index,
      columnConfig.name,
      evt.target.value
    );
  };
  const onkeyupHandler = (evt) => {
    if (!columnConfig.cellRenderer.onKeystroke) return;
    columnConfig.cellRenderer.onKeystroke(
      index,
      columnConfig.name,
      evt.target.value
    );
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
  $: validation = columnConfig.cellRenderer.validation
    ? columnConfig.cellRenderer.validation(row, row[columnConfig.name])
    : ValidationState.OK;

  const enum ValidationIcon {
    ERROR,
    WARNING,
    NONE,
  }

  let validationErrorMsg: string = undefined;
  let icon = ValidationIcon.NONE;
  $: if (validation !== ValidationState.OK) {
    // FIXME: for now, if a row has an invalid state, we know it is a MeasureDefinitionEntity, but this is not very robust
    // FIXME: currently, the `expressionValidationError.message` is only ever "Unexpected end of input"
    // We'll use a placeholder until we can get more detailed feedback
    // validationErrorMsg = (<MeasureDefinitionEntity>row)
    //   .expressionValidationError.message;
    "asdfad".trim;
    if (value.trim() === "") {
      validationErrorMsg = "This aggregation expression is empty";
    } else {
      validationErrorMsg = "This aggregation expression is invalid";
    }

    if (
      validation === ValidationState.ERROR &&
      editing === false &&
      value.trim() !== ""
    ) {
      // only show an error icon if not currently editing
      // and if there is actually a value in the input
      icon = ValidationIcon.ERROR;
    } else if (
      (validation === ValidationState.ERROR && editing === true) ||
      (validation === ValidationState.ERROR && value.trim() === "") ||
      validation === ValidationState.WARNING
    ) {
      icon = ValidationIcon.WARNING;
    } else {
      icon = ValidationIcon.NONE;
    }
  }
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
      autocomplete="off"
      bind:this={inputElt}
      id="model-title-input"
      class="table-input w-full text-ellipsis bg-inherit font-normal"
      on:input={startEditing}
      on:focus={startEditing}
      on:blur={stopEditing}
      on:change={onchangeHandler}
      on:keyup={onkeyupHandler}
      value={value ?? ""}
    />

    {#if icon !== ValidationIcon.NONE}
      <Tooltip location="top" alignment="middle" distance={16}>
        <div class="self-center" style="height:0px">
          <div style="position:relative; top:-10px;">
            {#if icon === ValidationIcon.ERROR}
              <!-- NOTE: #ef4444 === fill-red-500 -->
              <AlertCircle size={"20px"} color={"#ef4444"} />
            {:else if icon === ValidationIcon.WARNING}
              <!-- NOTE: #ca8a04 === fill-yellow-500 -->
              <AlertTriangle size={"20px"} color={"#ca8a04"} />
            {/if}
          </div>
        </div>

        <TooltipContent slot="tooltip-content">
          {validationErrorMsg}
        </TooltipContent>
      </Tooltip>
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
