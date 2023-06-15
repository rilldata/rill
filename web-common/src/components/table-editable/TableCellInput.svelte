<script lang="ts">
  import AlertCircle from "@rilldata/web-common/components/icons/AlertCircle.svelte";
  import AlertTriangle from "@rilldata/web-common/components/icons/AlertTriangle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ValidationState } from "@rilldata/web-common/features/metrics-views/errors";
  import type { EntityRecord } from "@rilldata/web-local/lib/temp/entity";
  import type {
    CellConfigInput,
    ColumnConfig,
    InputValidation,
  } from "./ColumnConfig";

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
  let validation: InputValidation;
  $: validation = columnConfig?.cellRenderer?.getInputValidation
    ? columnConfig.cellRenderer.getInputValidation(row, row[columnConfig.name])
    : { state: ValidationState.OK, message: "" };

  const enum ValidationIcon {
    ERROR,
    WARNING,
    NONE,
  }

  let validationErrorMsg: string = undefined;
  let icon = ValidationIcon.NONE;
  $: if (validation.state !== ValidationState.OK) {
    if (value.trim() === "") {
      validationErrorMsg = "This aggregation expression is empty";
    } else {
      validationErrorMsg = validation.message;
    }

    if (
      validation.state === ValidationState.ERROR &&
      editing === false &&
      value.trim() !== ""
    ) {
      // only show an error icon if not currently editing
      // and if there is actually a value in the input
      icon = ValidationIcon.ERROR;
    } else if (
      (validation.state === ValidationState.ERROR && editing === true) ||
      (validation.state === ValidationState.ERROR && value.trim() === "") ||
      validation.state === ValidationState.WARNING
    ) {
      icon = ValidationIcon.WARNING;
    } else {
      icon = ValidationIcon.NONE;
    }
  } else {
    icon = ValidationIcon.NONE;
  }
</script>

<td
  class="py-2 px-4 border border-gray-200 hover:bg-gray-200"
  on:click={startEditing}
  on:keydown={() => {
    /* prevent default behavior */
  }}
  style={editing
    ? "cursor:text; background:white; outline:1px solid #aaa;"
    : ""}
>
  <div class="flex flex-row">
    <input
      aria-label={columnConfig.ariaLabel}
      autocomplete="off"
      bind:this={inputElt}
      class={"table-input w-full text-ellipsis bg-inherit " +
        (columnConfig?.customClass || "")}
      id="model-title-input"
      on:blur={stopEditing}
      on:change={onchangeHandler}
      on:focus={startEditing}
      on:input={startEditing}
      on:keyup={onkeyupHandler}
      value={value ?? ""}
    />

    {#if icon !== ValidationIcon.NONE}
      <Tooltip location="top" alignment="middle" distance={16}>
        <div class="self-center z-10" style="height:0px">
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
