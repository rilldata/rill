<script lang="ts">
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import ContainerTitle from "../core/ContainerTitle.svelte";
  import RowContainer from "../core/RowContainer.svelte";
  import Measure from "./Measure.svelte";
  interface Measure {
    displayName?: string;
    expression?: string;
    description?: string;
    id: string;
    visible: boolean;
  }

  let showError = false;
  let showState = false;
  let errorGUID = guidGenerator();

  let measures: Measure[] = [
    {
      displayName: "Total Records",
      expression: "count(*)",
      id: guidGenerator(),
      visible: true,
    },
    ...Array.from({ length: 0 }).map((_, i) => {
      return {
        displayName: "measure" + i,
        expression: `count(${i})`,
        id: guidGenerator(),
        visible: true,
      };
    }),
    {
      displayName: "Revenue",
      expression: "sum(revenue)",
      id: errorGUID,
      visible: true,
    },
    {
      displayName: "Distinct Users",
      expression: "count (distinct user_id)",
      id: guidGenerator(),
      visible: true,
    },
    {
      displayName: "Distinct Non-Bot Users",
      expression: "count (distinct sanitized_user_id)",
      id: guidGenerator(),
      visible: false,
    },
    {
      displayName: "Total Sales",
      expression: "sum(sales_price)",
      id: guidGenerator(),
      visible: true,
    },
  ];
</script>

<input type="checkbox" bind:checked={showError} />
<input type="checkbox" bind:checked={showState} />

<!-- RowContainer keeps a copy of the top-level items,
  handles their manipulation, and then on edit,
  dispatches to update the container.
-->

<ContainerTitle>Measures</ContainerTitle>

<div>
  <div>Expression</div>
  <div>Display Name</div>
  <div>Description</div>
  <div>Format Preset</div>
</div>
<RowContainer
  items={measures}
  addItemText="Add Measure"
  on:update-items={(event) => {
    measures = event.detail;
  }}
  let:item
  let:edit
  let:moveUp
  let:moveDown
  let:moveToTop
  let:moveToBottom
  let:deleteItem
  let:select
  let:toggleVisibility
  let:selected
  let:mode
  let:isDragging
  let:dragHandleMousedown
>
  <Measure
    error={item.id === errorGUID && showError
      ? "This is an error message"
      : undefined}
    expression={item.expression}
    visible={item.visible}
    displayName={item.displayName}
    description={item.description}
    {selected}
    {mode}
    {isDragging}
    on:draghandle-mousedown={dragHandleMousedown}
    on:select={select}
    on:toggle-visibility={toggleVisibility}
    on:edit={edit}
    on:delete={deleteItem}
    on:move-up={moveUp}
    on:move-down={moveDown}
    on:move-to-top={moveToTop}
    on:move-to-bottom={moveToBottom}
  />
</RowContainer>
{#if showState}
  <pre>
    {JSON.stringify(measures, null, 2)}
  </pre>
{/if}
