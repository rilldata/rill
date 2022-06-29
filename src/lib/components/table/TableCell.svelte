<script lang="ts">
  /**
   * TableCell.svelte
   * notes:
   * - the max cell-width that preserves a timestamp is 210px.
   */
  import { createEventDispatcher } from "svelte";
  import { FormattedDataType } from "$lib/components/data-types/";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import ErrorIcon from "$lib/components/icons/CrossIcon.svelte";
  import WarningIcon from "$lib/components/icons/WarningIcon.svelte";
  import CopyableTableCell from "$lib/components/table/CopyableTableCell.svelte";
  import type { ColumnConfig } from "$lib/components/table/ColumnConfig";

  export let value;
  export let column: ColumnConfig;
  export let index = undefined;
  export let isNull = false;
  export let validation: ValidationState;

  let renderer: unknown;
  $: if (column) renderer = column.renderer ?? CopyableTableCell;

  const dispatch = createEventDispatcher();

  let activeCell = false;
</script>

<Tooltip location="top" distance={16}>
  <td
    on:mouseover={() => {
      dispatch("inspect", index);
      activeCell = true;
    }}
    on:mouseout={() => {
      activeCell = false;
    }}
    on:focus={() => {
      dispatch("inspect", index);
      activeCell = true;
    }}
    on:blur={() => {
      activeCell = false;
    }}
    title={value}
    class="
        p-2
        pl-4
        pr-4
        border
        border-gray-200
        {activeCell && 'bg-gray-200'}
    "
    style:width="var(--table-column-width-{column.name}, 210px)"
    style:max-width="var(--table-column-width-{column.name}, 210px)"
  >
    <svelte:component
      this={renderer}
      on:change={(evt) => dispatch("change", evt.detail)}
      {value}
      {isNull}
      {index}
      {column}
    />
    {#if validation === ValidationState.ERROR}
      <ErrorIcon className="inline" />
    {:else if validation === ValidationState.WARNING}
      <WarningIcon className="inline" />
    {/if}
  </td>
  {#if column.copyable}
    <TooltipContent slot="tooltip-content">
      <TooltipTitle>
        <svelte:fragment slot="name">
          <FormattedDataType {value} type={column.type} dark />
        </svelte:fragment>
      </TooltipTitle>
      <TooltipShortcutContainer>
        <div>
          <StackingWord>copy</StackingWord> this value to clipboard
        </div>
        <Shortcut>
          <span style="font-family: var(--system);">⇧</span> + Click
        </Shortcut>
      </TooltipShortcutContainer>
    </TooltipContent>
  {/if}
</Tooltip>
